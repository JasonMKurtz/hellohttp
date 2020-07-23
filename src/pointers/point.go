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

type Seeker int

const (
	top Seeker = iota
	prev
	next
	bottom
)

func (s Seeker) IsValid() bool {
	switch s {
	case top, prev, next, bottom:
		return true
	}
	return false
}

func CreateFeed() *Feed {
	return &Feed{}
}

func (f *Feed) Size() int {
	return f.length
}

func (f *Feed) CanDelete() bool {
	return f.Size() <= 0
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

func (f *Feed) Delete(title string) {
	for _, p := range f.List() {
		if p.title != title {
			continue
		}

		next := p.next
		prev := p.prev
		f.length--
		if next == nil && prev == nil { // deleting our only element
			f.topPost = nil
			f.curPost = nil
		} else if next == nil { // bottom of the list
			prev.next = nil
		} else if prev == nil { // top of the list
			next.prev = nil
			f.topPost = next
		} else { // some middle element
			p.prev.next = p.next
			p.next.prev = p.prev
		}
	}
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

func (f *Feed) seek(direction Seeker) {
	if !direction.IsValid() {
		return
	}

	switch direction {
	case top:
		f.curPost = f.topPost
		break
	case prev:
		prev := f.curPost.prev
		if prev == nil {
			return
		}
		f.curPost = prev
		break
	case next:
		next := f.curPost.next
		if next == nil {
			return
		}
		f.curPost = next
		break
	case bottom:
		f.curPost = f.Tail()
		break
	}
}

func (f *Feed) SeekPrev() {
	f.seek(prev)
}

func (f *Feed) SeekNext() {
	f.seek(next)
}

func (f *Feed) SeekHead() {
	f.seek(top)
}

func (f *Feed) SeekTail() {
	f.seek(bottom)
}

func (f *Feed) Top() *Post {
	return f.topPost
}

func (f *Feed) Current() *Post {
	return f.curPost
}

func (f *Feed) Tail() *Post {
	last := f.curPost
	for last.next != nil {
		last = last.next
	}

	return last
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
	for _, p := range f.List() {
		fmt.Printf("%s\n", ReadPost(p))
	}
}
