package issue1

import "time"

type (
	// Channel represents a singular stream of posts that a user can subscribe to
	// under administration by certain users.
	Channel struct {
		Username           string
		Name               string
		Description        string
		Owner              string
		Admins             []string
		PostIDs            []uint
		StickiedPostIDs    [2]uint
		OfficialReleaseIDs []uint
		ReleaseIDs         []uint
		CreationTime       time.Time
	}
	// Comment represents standard comments users can attach
	// to a post or another comment.
	// replyTo is either and id of another comment or -1 if
	// it's a reply to original post.
	Comment struct {
		ID           uint
		OriginPost   uint
		Commenter    string
		Content      string
		ReplyTo      uint
		CreationTime time.Time
	}
	// Feed is a value object that tracks channels that a user subbed to
	// and other settings
	Feed struct {
		ID            int         `json:"id,omitempty"`
		OwnerUsername string      `json:"ownerUsername"`
		Sorting       FeedSorting `json:"defaultSorting"`
		//Subscriptions []*Channel `json:"subscriptions"`
		// hiddenPosts   []Post
	}
	// Post is an aggregate entity of Releases along with socially interactive
	// components such as stars, posting user and comments attached to the post
	Post struct {
		ID            uint
		Title         string
		Description   string
		OriginChannel string
		Poster        string
		ContentIDs    []uint
		Stars         map[string]uint // map of a username to the number of stars (range of 0 to 5) given
		CreationTime  time.Time
	}
	// Release represents an atomic work of creativity.
	Release struct {
		ID           uint
		OwnerChannel string
		ContentType  string
		Content      string // if text: text. if image: url
		Metadata     Metadata
		CreationTime time.Time
	}
	// Metadata is a value object holds all the metadata of releases.
	// genreDefining is the genre classification that defines the release most.
	// authors contains username in string form if author is an issue#1 user
	// or plain names otherwise.
	// description is for data like blurb.
	Metadata struct {
		Authors       []string `json:"authors"`
		GenreDefining string   `json:"genreDefining,omitempty"`
		Genres        []string `json:"genres"`
		Description   string   `json:"description,omitempty"`
	}
	// User represents standard user entity of issue#1.
	// bookmarkedPosts map contains the postId mapped to the time it was bookmarked.
	User struct {
		Username     string    `json:"username"`
		Email        string    `json:"email"`
		FirstName    string    `json:"firstName"`
		MiddleName   string    `json:"middleName"`
		LastName     string    `json:"lastName"`
		CreationTime time.Time `json:"creationTime"`
		Bio          string    `json:"bio"`
		//BookmarkedPosts map[time.Time]Post `json:"bookmarkedPosts"`
		Password   string `json:"password,omitempty"`
		PictureURL string `json:"pictureURL"`
	}
)
