package main

import (
	"context"
	"fmt"
	redissrv "github.com/redis-go/redis"
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
	go log.Fatal(redissrv.Run(":6379"))

	// Initialize Redis client
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// Check Redis connection
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
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
		err := redisClient.HMSet(ctx, itemID, map[string]interface{}{
			"name":        name,
			"price":       price,
			"description": description,
		}).Err()
		if err != nil {
			http.Error(w, "Error adding item", http.StatusInternalServerError)
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

		// Remove the item
		err := redisClient.Del(ctx, itemID).Err()
		if err != nil {
			http.Error(w, "Error removing item", http.StatusInternalServerError)
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
		err := redisClient.HMSet(ctx, categoryID, map[string]interface{}{
			"name": name,
		}).Err()
		if err != nil {
			http.Error(w, "Error adding category", http.StatusInternalServerError)
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

		// Remove the category
		err := redisClient.Del(ctx, categoryID).Err()
		if err != nil {
			http.Error(w, "Error removing category", http.StatusInternalServerError)
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

		// Associate the item with the category
		err := redisClient.SAdd(ctx, categoryID+":items", itemID).Err()
		if err != nil {
			http.Error(w, "Error associating item with category", http.StatusInternalServerError)
			return
		}

		// Redirect to the index page
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func disassociateItemCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Parse form values
		itemID := r.FormValue("item_id")
		categoryID := r.FormValue("category_id")

		// Disassociate the item from the category
		err := redisClient.SRem(ctx, categoryID+":items", itemID).Err()
		if err != nil {
			http.Error(w, "Error disassociating item from category", http.StatusInternalServerError)
			return
		}

		// Redirect to the index page
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func fetchAllItems() ([]Item, error) {
	// Fetch all item keys
	keys, err := redisClient.Keys(ctx, "item:*").Result()
	if err != nil {
		return nil, err
	}

	// Fetch item details
	var items []Item
	for _, key := range keys {
		itemMap, err := redisClient.HGetAll(ctx, key).Result()
		if err != nil {
			return nil, err
		}
		price, _ := strconv.ParseFloat(itemMap["price"], 64)
		items = append(items, Item{
			ID:          key,
			Name:        itemMap["name"],
			Price:       price,
			Description: itemMap["description"],
		})
	}
	return items, nil
}

func fetchAllCategories() ([]Category, error) {
	// Fetch all category keys
	keys, err := redisClient.Keys(ctx, "category:*").Result()
	if err != nil {
		return nil, err
	}

	// Fetch category details
	var categories []Category
	for _, key := range keys {
		categoryMap, err := redisClient.HGetAll(ctx, key).Result()
		if err != nil {
			return nil, err
		}
		categories = append(categories, Category{
			ID:   key,
			Name: categoryMap["name"],
		})
	}
	return categories, nil
}
