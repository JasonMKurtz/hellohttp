package main

import "fmt"

type Feed struct {
	length  int
	topPost *Post
	curPost *Post
}

type Post struct {
	title string
	text  string
	next  *Post
	prev  *Post
}

func CreateFeed() *Feed {
	return &Feed{}
}

func (f *Feed) AddPost(title, text string) {
	p := &Post{
		title: title,
		text:  text,
	}

	if f.curPost == nil {
		f.topPost = p
		f.curPost = p
		fmt.Printf("Added \"%s\" as the first post.\n", p.title)
		return
	}

	cur := f.curPost
	for cur.next != nil {
		cur = cur.next
	}

	fmt.Printf("We're adding \"%s\". The previous one is \"%s\"\n",
		p.title, cur.title)

	fmt.Printf("Cur: %s\n", cur.title)

	p.prev = cur
	cur.next = p
	f.curPost = p
	return
}

func (f *Feed) DelPost(title string) {
}

func (f *Feed) List() []Post {
	if f.curPost == nil {
		return nil
	}
	var posts []Post

	cur := f.curPost
	posts = append(posts, *cur)
	for cur.next != nil {
		posts = append(posts, *cur.next)
		cur = cur.next
	}

	return posts
}

func (f *Feed) Tail() Post {
	last := f.curPost
	for last.next != nil {
		last = last.next
	}

	return *last
}

func (f *Feed) seek(direction int) {
	if direction == 0 {
		prev := f.curPost.prev
		if prev == nil {
			return
		}
		*f.curPost = *prev
	} else if direction == 1 {
		next := f.curPost.next
		if next == nil {
			return
		}
		*f.curPost = *next
	}
}

func (f *Feed) Prev() Post {
	f.seek(0)
	return *f.curPost
}

func (f *Feed) Next() Post {
	f.seek(1)
	return *f.curPost
}

func (f *Feed) Top() *Post {
	return f.topPost
}

func (f *Feed) Current() *Post {
	return f.curPost
}

func ReadPost(p *Post) string {
	next := p.next
	nextTitle := ""
	prevTitle := ""
	if next != nil {
		nextTitle = next.title
	}

	prev := p.prev
	if prev != nil {
		prevTitle = prev.title
	}
	return fmt.Sprintf(
		"\"%s\" -> %s\n"+
			"Next: %s\nPrev: %s\n\n",
		p.title,
		p.text,
		nextTitle, prevTitle,
	)
}

func main() {
	f := CreateFeed()
	f.AddPost("test post", "lorem")
	f.AddPost("test post 2", "lorem")
	fmt.Printf(ReadPost(f.Current()))
	f.Prev()
	fmt.Printf(ReadPost(f.Current()))
	f.Next()
	fmt.Printf(ReadPost(f.Current()))
}
