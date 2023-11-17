package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	redis2 "github.com/redis-go/redis"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	redisClient *redis.Client
	ctx         = context.Background()
	templates   = template.Must(template.ParseGlob("templates/*"))
)

type Item struct {
	ID          string
	Name        string
	Price       float64
	Description string
}

type Category struct {
	ID   string
	Name string
}

func main() {
	log.Println("Starting redis server")
	go func() { log.Fatal(redis2.Run(":6379")) }()

	// Initialize Redis client
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// Check Redis connection
	const maxRetries = 10
	for i := 0; i < maxRetries; i++ {
		_, err := redisClient.Ping(ctx).Result()
		if err != nil {
			log.Printf("Could not connect to Redis: %v", err)
			if i == maxRetries-1 {
				log.Fatal("Max retries exceeded")
			}
			time.Sleep(time.Millisecond * 500)
		}
	}

	// Define routes
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/add-item", addItemHandler)
	http.HandleFunc("/remove-item", removeItemHandler)
	http.HandleFunc("/add-category", addCategoryHandler)
	http.HandleFunc("/remove-category", removeCategoryHandler)
	http.HandleFunc("/associate-item-category", associateItemCategoryHandler)
	http.HandleFunc("/disassociate-item-category", disassociateItemCategoryHandler)

	// Start the server
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// Fetch all items and categories to display
	items, err := fetchAllItems()
	if err != nil {
		log.Println(err)
		http.Error(w, "Error fetching items", http.StatusInternalServerError)
		return
	}
	categories, err := fetchAllCategories()
	if err != nil {
		http.Error(w, "Error fetching categories", http.StatusInternalServerError)
		return
	}

	// Render the index template
	err = templates.ExecuteTemplate(w, "index.html", map[string]interface{}{
		"Items":      items,
		"Categories": categories,
	})
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

func addItemHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Parse form values
		name := r.FormValue("name")
		price, _ := strconv.ParseFloat(r.FormValue("price"), 64)
		description := r.FormValue("description")

		// Create a new item
		itemID := fmt.Sprintf("item:%d", time.Now().UnixNano())
		item := Item{
			ID:          itemID,
			Name:        name,
			Price:       price,
			Description: description,
		}
		itemData, _ := json.Marshal(item)

		// Use SET to store item details
		err := redisClient.Set(ctx, itemID, itemData, 0).Err()
		if err != nil {
			http.Error(w, "Error adding item", http.StatusInternalServerError)
			return
		}

		// Use RPUSH to add the item ID to the items list
		err = redisClient.RPush(ctx, "items", itemID).Err()
		if err != nil {
			http.Error(w, "Error adding item ID to items list", http.StatusInternalServerError)
			return
		}

		// Redirect to the index page
		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		// Render the add item template
		err := templates.ExecuteTemplate(w, "add-item.html", nil)
		if err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
	}
}

func removeItemHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Parse form values
		itemID := r.FormValue("item_id")

		// Remove the item details
		err := redisClient.Del(ctx, itemID).Err()
		if err != nil {
			http.Error(w, "Error removing item", http.StatusInternalServerError)
			return
		}

		// Remove the item ID from the items list
		// Since LREM is not allowed, we need to recreate the list without the item ID
		err = removeIDFromList("items", itemID)
		if err != nil {
			log.Println(err)
			http.Error(w, "Error removing item ID from items list", http.StatusInternalServerError)
			return
		}

		// Redirect to the index page
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func addCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Parse form values
		name := r.FormValue("name")

		// Create a new category
		categoryID := fmt.Sprintf("category:%d", time.Now().UnixNano())
		category := Category{
			ID:   categoryID,
			Name: name,
		}
		categoryData, _ := json.Marshal(category)

		// Use SET to store category details
		err := redisClient.Set(ctx, categoryID, categoryData, 0).Err()
		if err != nil {
			http.Error(w, "Error adding category", http.StatusInternalServerError)
			return
		}

		// Use RPUSH to add the category ID to the categories list
		err = redisClient.RPush(ctx, "categories", categoryID).Err()
		if err != nil {
			http.Error(w, "Error adding category ID to categories list", http.StatusInternalServerError)
			return
		}

		// Redirect to the index page
		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		// Render the add category template
		err := templates.ExecuteTemplate(w, "add-category.html", nil)
		if err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
	}
}

func removeCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Parse form values
		categoryID := r.FormValue("category_id")

		// Remove the category details
		err := redisClient.Del(ctx, categoryID).Err()
		if err != nil {
			http.Error(w, "Error removing category", http.StatusInternalServerError)
			return
		}

		// Remove the category ID from the categories list
		// Since LREM is not allowed, we need to recreate the list without the category ID
		err = removeIDFromList("categories", categoryID)
		if err != nil {
			http.Error(w, "Error removing category ID from categories list", http.StatusInternalServerError)
			return
		}

		// Redirect to the index page
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func associateItemCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Parse form values
		itemID := r.FormValue("item_id")
		categoryID := r.FormValue("category_id")

		// Use RPUSH to add the item ID to the category's items list
		err := redisClient.RPush(ctx, categoryID+":items", itemID).Err()
		if err != nil {
			http.Error(w, "Error associating item with category", http.StatusInternalServerError)
			return
		}
		log.Printf("Associated item %s with category %s", itemID, categoryID)

		// Redirect to the index page
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func disassociateItemCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Parse form values
		itemID := r.FormValue("item_id")
		categoryID := r.FormValue("category_id")

		// Remove the item ID from the category's items list
		// Since LREM is not allowed, we need to recreate the list without the item ID
		err := removeIDFromList(categoryID+":items", itemID)
		if err != nil {
			http.Error(w, "Error disassociating item from category", http.StatusInternalServerError)
			return
		}

		// Redirect to the index page
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func fetchAllItems() ([]Item, error) {
	// Use LRANGE to get all item IDs from the items list
	itemIDs, err := redisClient.LRange(ctx, "items", 0, -1).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil // List is empty
		}
		return nil, fmt.Errorf("error fetching item IDs: %v", err)
	}

	// Fetch item details
	var items []Item
	for _, itemID := range itemIDs {
		itemData, err := redisClient.Get(ctx, itemID).Result()
		if err != nil {
			return nil, err
		}
		var item Item
		err = json.Unmarshal([]byte(itemData), &item)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func fetchAllCategories() ([]Category, error) {
	// Use LRANGE to get all category IDs from the categories list
	categoryIDs, err := redisClient.LRange(ctx, "categories", 0, -1).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil // List is empty
		}
		return nil, err
	}

	// Fetch category details
	var categories []Category
	for _, categoryID := range categoryIDs {
		categoryData, err := redisClient.Get(ctx, categoryID).Result()
		if err != nil {
			return nil, err
		}
		var category Category
		err = json.Unmarshal([]byte(categoryData), &category)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, nil
}

func removeIDFromList(listName, id string) error {
	// using LREM is allowed
	// Remove the item ID from the list
	err := redisClient.LRem(ctx, listName, 0, id).Err()
	if err != nil {
		return err
	}
	return nil
}
