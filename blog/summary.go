package blog

type Category struct {
	Key string
	Blogs
}

type Summary []Category
