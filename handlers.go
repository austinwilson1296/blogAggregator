package main

import (
	"context"
	"fmt"
	"time"
	"github.com/google/uuid"
	"github.com/austinwilson1296/blogaggregator/internal/database"
	"log"
	"strconv"
)

func handlerUsers(s *state,cmd command)error{
	if len(cmd.Args) != 0 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}
	users,err := s.db.GetUsers(context.Background())
	if err != nil{
		return fmt.Errorf("error reseting db: %w",err)
	}
	
	for _,user := range users{
		if user.Name == s.cfg.CurrentUserName{
			fmt.Printf("* %s (current)\n",user.Name)
		}
		fmt.Printf("* %s\n",user.Name)
	}
	return nil
}

func handlerReset(s *state,cmd command) error{
	if len(cmd.Args) != 0 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}

	err := s.db.DeleteUsers(context.Background())
	if err != nil{
		return fmt.Errorf("error reseting db: %w",err)
	}
	fmt.Println("Database reset complete")
	return nil
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}
	name := cmd.Args[0]

	_, err := s.db.GetUser(context.Background(), name)
	if err != nil {
		return fmt.Errorf("couldn't find user: %w", err)
	}

	err = s.cfg.SetUser(name)
	if err != nil {
		return fmt.Errorf("couldn't set current user: %w", err)
	}

	fmt.Println("User switched successfully!")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %v <name>", cmd.Name)
	}

	name := cmd.Args[0]

	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      name,
	})
	if err != nil {
		return fmt.Errorf("couldn't create user: %w", err)
	}

	err = s.cfg.SetUser(user.Name)
	if err != nil {
		return fmt.Errorf("couldn't set current user: %w", err)
	}

	fmt.Println("User created successfully:")
	printUser(user)
	return nil
}

func printUser(user database.User) {
	fmt.Printf(" * ID:      %v\n", user.ID)
	fmt.Printf(" * Name:    %v\n", user.Name)
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.Args) < 1 || len(cmd.Args) > 2 {
		return fmt.Errorf("usage: %v <time_between_reqs>", cmd.Name)
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}

	log.Printf("Collecting feeds every %s...", timeBetweenRequests)

	ticker := time.NewTicker(timeBetweenRequests)

	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}
func handlerAddFeed(s *state, cmd command, user database.User) error {

	if len(cmd.Args) != 2 {
		return fmt.Errorf("usage: %s <name> <url>", cmd.Name)
	}

	name := cmd.Args[0]
	url := cmd.Args[1]

	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		Name:      name,
		Url:       url,
	})
	if err != nil {
		return fmt.Errorf("couldn't create feed: %w", err)
	}
	_,err = s.db.CreateFeedFollow(context.Background(),database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})

	fmt.Println("Feed created successfully:")
	printFeed(feed)
	fmt.Println()
	fmt.Println("=====================================")

	return nil
}

func printFeed(feed database.Feed) {
	fmt.Printf("* ID:            %s\n", feed.ID)
	fmt.Printf("* Created:       %v\n", feed.CreatedAt)
	fmt.Printf("* Updated:       %v\n", feed.UpdatedAt)
	fmt.Printf("* Name:          %s\n", feed.Name)
	fmt.Printf("* URL:           %s\n", feed.Url)
	fmt.Printf("* UserID:        %s\n", feed.UserID)
}

func handlerFeeds(s *state, cmd command) error{
	if len(cmd.Args) != 0 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}
	feeds,err := s.db.GetFeeds(context.Background())
	if err != nil{
		return fmt.Errorf("error retrieving feeds %w",err)
	}
	for _,feed := range feeds{
		fmt.Printf("feed name: %s\nfeed url %s\nuser name %s\n",feed.Name,feed.Url,feed.Name_2)
	}
	return nil
}

func handlerFollow(s *state,cmd command, user database.User) error{

	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}
	url := cmd.Args[0]
	feedName,err := s.db.GetFeedByURL(context.Background(),url)
	if err != nil{
		return fmt.Errorf("no feed found for given URL: %w",err)
	}

	_, err = s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    feedName.ID,
	})
	if err != nil{
		return fmt.Errorf("unable to follow feed: %s",feedName)
	}
	fmt.Println("Follow Succesful!")
	fmt.Printf("User: %s\nFeed:%s\n",user.Name,feedName.Name)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
    // user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
    // if err != nil {
    //     return err
    // }
    
    feeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
    if err != nil {
        return err
    }
	fmt.Printf("User: %s\n",user.Name)
	fmt.Println("Following:")
    for _,feed := range feeds{
		fmt.Printf("%s\n",feed)
	}
    
    return nil
}

func handlerUnfollow(s *state,cmd command, user database.User)error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}
	url := cmd.Args[0]
	feedURL,err := s.db.GetFeedByURL(context.Background(),url)
	if err != nil{
		return fmt.Errorf("no feed found for given URL: %w",err)
	}
	err = s.db.UnfollowFeed(context.Background(),database.UnfollowFeedParams{
		Url: url,
		UserID: user.ID,
	})
	if err != nil{
		return err
	}
	fmt.Printf("%s Unfollowed",feedURL)
	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	limit := 2
	if len(cmd.Args) == 1 {
		if specifiedLimit, err := strconv.Atoi(cmd.Args[0]); err == nil {
			limit = specifiedLimit
		} else {
			return fmt.Errorf("invalid limit: %w", err)
		}
	}

	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		return fmt.Errorf("couldn't get posts for user: %w", err)
	}

	fmt.Printf("Found %d posts for user %s:\n", len(posts), user.Name)
	for _, post := range posts {
		fmt.Printf("%s from %s\n", post.PublishedAt.Time.Format("Mon Jan 2"), post.FeedName)
		fmt.Printf("--- %s ---\n", post.Title)
		fmt.Printf("    %v\n", post.Description.String)
		fmt.Printf("Link: %s\n", post.Url)
		fmt.Println("=====================================")
	}

	return nil
}