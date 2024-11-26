package main

import (
	"context"
	"fmt"
	"time"
	"github.com/google/uuid"
	"github.com/austinwilson1296/blogaggregator/internal/database"

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
	feed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return fmt.Errorf("error fetching feed in handlerAgg: %w", err)
	}

	fmt.Printf("Channel Title: %s\n", feed.Channel.Title)
	fmt.Printf("Channel Description: %s\n", feed.Channel.Description)
	for _, item := range feed.Channel.Item {
		fmt.Printf("Item Title: %s\n", item.Title)
		fmt.Printf("Item Description: %s\n", item.Description)
		fmt.Printf("Item PubDate: %s\n", item.PubDate)
	}

	return nil
}