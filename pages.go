package main

import (
	"html/template"
	"net/http"
	"strconv"
)

type Data struct {
	// Template header
	LoggedIn bool
	User     User
	// Error messages for form validation
	Message *Message
	// Template data (for different pages)
	Posts []Post
	Post  Post
	// All threads for search purposes
	Threads []string
	// Is signin modal open
	SigninModalOpen string
	// Is signup modal open
	SignupModalOpen string
	// Scroll page to post
	ScrollTo string
	// saves current filter
	Filter string

	Notifications []Notification
}

type ErrorMsg struct {
	Status  int
	Message string
}

// How to get rid of this?
var errMsg ErrorMsg

func index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		createError(w, r, http.StatusNotFound)
		return
	}

	setLastPage(w, r.URL.Path)

	// get data for index page
	data := welcome(w, r)

	if r.URL.Query().Get("modal") != "" {
		data.SigninModalOpen = r.URL.Query().Get("modal")
	}

	posts := fetchAllPosts(database)
	reverse(posts)

	data.Filter = r.URL.Query().Get("filter")
	if data.Filter == "" || data.Filter == "All Categories" || contains(data.Threads, data.Filter) {
		if contains(data.Threads, data.Filter) {
			posts = filterByThread(posts, data.Filter)
		}
	} else {
		createError(w, r, http.StatusBadRequest)
		return
	}

	data.Posts = fillPosts(&data, posts)

	tmpl, err := template.ParseFiles("static/template/index.html", "static/template/base.html")
	if err != nil {
		createError(w, r, http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		createError(w, r, http.StatusInternalServerError)
		return
	}
}

func post(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	posts := fetchAllPosts(database)
	if id > 0 && id <= len(posts) {
		data := welcome(w, r)
		data.Post = fetchPostByID(database, id)
		data.Post.Comments = fetchCommentsByPost(database, id)
		if data.LoggedIn {
			data.Post.UserReaction = fetchReactionByUserAndId(database, "postsReactions", data.User.Id, data.Post.Id).Value
		}

		for i := 0; i < len(data.Post.Comments); i++ {
			data.Post.Comments[i].User = fetchUserById(database, data.Post.Comments[i].UserId)
			if data.LoggedIn {
				data.Post.Comments[i].UserReaction = fetchReactionByUserAndId(database, "commentsReactions", data.User.Id, data.Post.Comments[i].Id).Value
			}
		}

		data.Post.User = fetchUserById(database, data.Post.UserId)

		setLastPage(w, "/post/id?id="+strconv.Itoa(id))

		tmpl, err := template.ParseFiles("static/template/post.html", "static/template/base.html")
		if err != nil {
			createError(w, r, http.StatusInternalServerError)
			return
		}
		err = tmpl.Execute(w, data)
		if err != nil {
			createError(w, r, http.StatusInternalServerError)
			return
		}
	} else {
		createError(w, r, http.StatusBadRequest)
		return
	}
}

func commentedPosts(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/commentedPosts" {
		createError(w, r, http.StatusNotFound)
		return
	}
	setLastPage(w, r.URL.Path)

	// get all commented posts and their comments
	data := welcome(w, r)

	posts := fetchPostsByUserComments(database, data.User.Id)

	data.Filter = r.URL.Query().Get("filter")
	if data.Filter == "" || data.Filter == "All Categories" || contains(data.Threads, data.Filter) {
		if contains(data.Threads, data.Filter) {
			posts = filterByThread(posts, data.Filter)
		}
	} else {
		createError(w, r, http.StatusBadRequest)
		return
	}

	data.Posts = fillPosts(&data, posts)
	reverse(posts)

	tmpl, err := template.ParseFiles("static/template/commentedPosts.html", "static/template/base.html")
	if err != nil {
		createError(w, r, http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, data)
	if err != nil {
		createError(w, r, http.StatusInternalServerError)
		return
	}
}

func dashBoard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/dashBoard" {
		createError(w, r, http.StatusNotFound)
		return
	}

	data := welcome(w, r)
	tmpl, err := template.ParseFiles("static/template/dashBoard.html", "static/template/base.html")
	if err != nil {
		createError(w, r, http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, data)
	if err != nil {
		createError(w, r, http.StatusInternalServerError)
		return
	}
}

func myPosts(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/myPosts" {
		createError(w, r, http.StatusNotFound)
		return
	}
	setLastPage(w, r.URL.Path)

	// get data for index page
	data := welcome(w, r)

	posts := fetchPostsByUser(database, data.User.Id)

	data.Filter = r.URL.Query().Get("filter")
	if data.Filter == "" || data.Filter == "All Categories" || contains(data.Threads, data.Filter) {
		if contains(data.Threads, data.Filter) {
			posts = filterByThread(posts, data.Filter)
		}
	} else {
		createError(w, r, http.StatusBadRequest)
		return
	}

	data.Posts = fillPosts(&data, posts)
	reverse(posts)

	tmpl, err := template.ParseFiles("static/template/myPosts.html", "static/template/base.html")
	if err != nil {
		createError(w, r, http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		createError(w, r, http.StatusInternalServerError)
		return
	}
}

func newPost(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/newPost" {
		createError(w, r, http.StatusNotFound)
		return
	}
	tmpl, err := template.ParseFiles("static/template/newPost.html", "static/template/base.html")
	if err != nil {
		createError(w, r, http.StatusInternalServerError)
		return
	}
	data := welcome(w, r)
	err = tmpl.Execute(w, data)
	if err != nil {
		createError(w, r, http.StatusInternalServerError)
		return
	}
}

// need to add logic to fetch liked posts
func likedPosts(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/likedPosts" {
		createError(w, r, http.StatusNotFound)
		return
	}
	setLastPage(w, r.URL.Path)

	data := welcome(w, r)
	posts := fetchLikedPostsByUser(database, data.User.Id)

	data.Filter = r.URL.Query().Get("filter")
	if data.Filter == "" || data.Filter == "All Categories" || contains(data.Threads, data.Filter) {
		if contains(data.Threads, data.Filter) {
			posts = filterByThread(posts, data.Filter)
		}
	} else {
		createError(w, r, http.StatusBadRequest)
		return
	}

	data.Posts = fillPosts(&data, posts)
	reverse(posts)

	tmpl, err := template.ParseFiles("static/template/likedPosts.html", "static/template/base.html")
	if err != nil {
		createError(w, r, http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		createError(w, r, http.StatusInternalServerError)
		return
	}
}

// need to add logic to fetch disliked posts
func dislikedPosts(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/dislikedPosts" {
		createError(w, r, http.StatusNotFound)
		return
	}
	setLastPage(w, r.URL.Path)

	data := welcome(w, r)
	posts := fetchDislikedPostsByUser(database, data.User.Id)

	data.Filter = r.URL.Query().Get("filter")
	if data.Filter == "" || data.Filter == "All Categories" || contains(data.Threads, data.Filter) {
		if contains(data.Threads, data.Filter) {
			posts = filterByThread(posts, data.Filter)
		}
	} else {
		createError(w, r, http.StatusBadRequest)
		return
	}

	data.Posts = fillPosts(&data, posts)
	reverse(posts)

	tmpl, err := template.ParseFiles("static/template/dislikedPosts.html", "static/template/base.html")
	if err != nil {
		createError(w, r, http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		createError(w, r, http.StatusInternalServerError)
		return
	}
}

// func editComment(w http.ResponseWriter, r *http.Request) {
// 	if r.URL.Path != "/editComment" {
// 		createError(w, r, http.StatusNotFound)
// 		return
// 	}
// 	tmpl, err := template.ParseFiles("static/template/editComment.html")
// 	//	tmpl, err := template.ParseFiles("static/template/index.html", "static/template/base.html")
// 	if err != nil {
// 		createError(w, r, http.StatusInternalServerError)
// 		return
// 	}
// 	err = tmpl.Execute(w, nil)
// 	if err != nil {
// 		createError(w, r, http.StatusInternalServerError)
// 		return
// 	}
// }

// func editPost(w http.ResponseWriter, r *http.Request) {
// 	if r.URL.Path != "/editPost" {
// 		createError(w, r, http.StatusNotFound)
// 		return
// 	}
// 	tmpl, err := template.ParseFiles("static/template/editPost.html")
// 	//	tmpl, err := template.ParseFiles("static/template/index.html", "static/template/base.html")
// 	if err != nil {
// 		createError(w, r, http.StatusInternalServerError)
// 		return
// 	}
// 	err = tmpl.Execute(w, nil)
// 	if err != nil {
// 		return
// 	}
// }

func createError(w http.ResponseWriter, r *http.Request, status int) {
	// err := &ErrorMsg{}
	switch status {
	case 400:
		errMsg.Status = 400
		errMsg.Message = "Bad request"
	case 404:
		errMsg.Status = 404
		errMsg.Message = "Page not found"
	case 500:
		errMsg.Status = 500
		errMsg.Message = "Unable to execute the page"
	default:
		errMsg.Status = 418
		errMsg.Message = "Another error we even don't know about"
	}
	http.Redirect(w, r, "/error", http.StatusFound)
}

func showError(w http.ResponseWriter, r *http.Request) {
	errorTmpl, err := template.ParseFiles("static/template/error.html")

	// errorTmpl, err := template.ParseFiles("static/templates/error.html", "static/templates/base.html")
	if err != nil {
		http.Error(w, "Unable to parse error template", 500)
		return
	}

	err = errorTmpl.Execute(w, errMsg)
	if err != nil {
		http.Error(w, "Unable to execute error template", 500)
		return
	}
}

// if login true redirect from url /register and /login to main page
