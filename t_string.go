package redis

const StringType = uint64(0)
const StringTypeFancy = "string"

var _ Item = (*String)(nil)

type String struct {
	str *string
}

func NewString(str *string) *String {
	return &String{str: str}
}

func (s *String) Value() interface{} {
	return s.str
}

func (s *String) ValueType() uint64 {
	return StringType
}

func (s *String) ValueTypeFancy() string {
	return StringTypeFancy
}

func (s *String) OnDelete(key *string, db *RedisDb) {
}
