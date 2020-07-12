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

func (f *Feed) Size() int {
	return f.length
}

func (f *Feed) AddPost(title, text string) {
	p := Post{
		title: title,
		text:  text,
	}

	f.length++

	if f.curPost == nil {
		f.topPost = &p
		f.curPost = &p
		return
	}

	tail := f.Tail()
	p.prev = tail
	tail.next = &p
	f.curPost = &p
}

func (f *Feed) DelPost(title string) {
	for _, p := range f.List() {
		if p.title == title {
			p.prev.next = p.next
			f.curPost = p
		}
	}
	f.length--
}

func (f *Feed) List() []*Post {
	if f.curPost == nil {
		return nil
	}
	var posts []*Post

	cur := f.topPost
	posts = append(posts, cur)
	for cur.next != nil {
		posts = append(posts, cur.next)
		cur = cur.next
	}

	return posts
}

func (f *Feed) Tail() *Post {
	last := f.curPost
	for last.next != nil {
		last = last.next
	}

	return last
}

func (f *Feed) seek(direction int) {
	if direction == 0 {
		prev := f.curPost.prev
		if prev == nil {
			return
		}
		f.curPost = prev
	} else if direction == 1 {
		next := f.curPost.next
		if next == nil {
			return
		}
		f.curPost = next
	}
}

func (f *Feed) Prev() *Post {
	f.seek(0)
	return f.curPost
}

func (f *Feed) Next() *Post {
	f.seek(1)
	return f.curPost
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
	f.AddPost("test post 3", "lorem")
	f.AddPost("test post 4", "lorem")
	for _, p := range f.List() {
		fmt.Printf(ReadPost(p))
	}
}
