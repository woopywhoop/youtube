package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type UserData struct {
	mu               sync.RWMutex
	UserId           string `json:"uid"`
	DisplayName      string `json:"display_name"`
	GameLevel        int    `json:"game_level"`
	Experience       int64  `json:"experience"`
	UserInternalData string `json:"-"`
}

/*
-- ChatGPT prompt example that can generate protected getters and setters for struct fields

Be laconic and output only code
As a professional golang developer create get and set methods for ALL public fields of the following struct, utilizing mu mutex to make it thread safe

type UserData struct {
	mu               sync.RWMutex
	UserId           string `json:"uid"`
	DisplayName      string `json:"display_name"`
	GameLevel        int    `json:"game_level"`
	Experience       int64  `json:"experience"`
	UserInternalData string `json:"-"`
}
*/

func NewUserData(userId string, displayName string, gameLevel int, experience int64) *UserData {
	return &UserData{
		UserId:      userId,
		DisplayName: displayName,
		GameLevel:   gameLevel,
		Experience:  experience,
	}
}

func (u *UserData) GetDisplayName() string {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.DisplayName
}

func (u *UserData) SetDisplayName(displayName string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.DisplayName = displayName
}

func (u *UserData) GetGameLevel() int {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.GameLevel
}

func (u *UserData) SetGameLevel(gameLevel int) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.GameLevel = gameLevel
}

func (u *UserData) GetExperience() int64 {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.Experience
}

func (u *UserData) SetExperience(value int64) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.Experience = value
}

func (u *UserData) ToApi() string {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return MustStringify(u)
}

func MustStringify(obj interface{}) string {
	bytea, err := json.Marshal(obj)
	if err != nil {
		return ""
	}
	return string(bytea)
}

func (u *UserData) UpdateData(operation func(userdata *UserData)) {
	u.mu.Lock()
	defer u.mu.Unlock()
	operation(u)
}

type UsersCache struct {
	mu           sync.RWMutex
	userDataById map[string]*UserData
}

func NewUsersCache() *UsersCache {
	return &UsersCache{
		userDataById: make(map[string]*UserData),
	}
}

func (uc *UsersCache) GetUserData(userId string) (*UserData, bool) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	userData, found := uc.userDataById[userId]
	return userData, found
}

func (uc *UsersCache) AddUserData(users ...*UserData) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	for _, user := range users {
		uc.userDataById[user.UserId] = user
	}
}

// -- Example operations on cache

// Operation on each user data, thread safety of user data access managed by operation function

func (uc *UsersCache) PerformReadOperation(operation func(userData *UserData)) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	for _, userData := range uc.userDataById {
		operation(userData)
	}
}

func (uc *UsersCache) GetSafeCopySlice() []*UserData {
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	res := make([]*UserData, len(uc.userDataById))
	for _, userData := range uc.userDataById {
		res = append(res, userData)
	}
	return res
}

func (uc *UsersCache) MapReduceUsersWithFilter(
	filter func(userData *UserData) bool,
	mapper func(userData *UserData) interface{},
	reducer func([]interface{}) interface{},
) interface{} {
	uc.mu.RLock()
	defer uc.mu.RUnlock()

	// Map phase with filtering
	mappedResults := make([]interface{}, 0)
	for _, userData := range uc.userDataById {
		if filter(userData) {
			result := mapper(userData)
			mappedResults = append(mappedResults, result)
		}
	}

	// Reduce phase
	return reducer(mappedResults)
}

// Filter function to exclude users named "John"
func excludeJohnFilter(userData *UserData) bool {
	return userData.GetDisplayName() != "John"
}

// Mapper function to count users by level
func userLevelMapper(userData *UserData) interface{} {
	return map[string]int{
		"level": userData.GetGameLevel(),
		"count": 1,
	}
}

// Reducer function to combine counts for each level
func levelCountReducer(results []interface{}) interface{} {
	levelCounts := make(map[int]int)

	for _, result := range results {
		levelData := result.(map[string]int)
		level := levelData["level"]
		count := levelData["count"]

		levelCounts[level] += count
	}

	return levelCounts
}

func LoadUsersDataFromDB(usersCache *UsersCache) error {
	// Mock for actual implementation
	usersCache.AddUserData(
		NewUserData("uid_001", "king", 1, 100),
		NewUserData("uid_002", "queen", 1, 110),
		NewUserData("uid_003", "soldier", 1, 120),
		NewUserData("uid_004", "John", 1, 120),
	)
	return nil
}

func main() {
	usersCache := NewUsersCache()
	_ = LoadUsersDataFromDB(usersCache)
	for i := 0; i < 100; i++ {
		// iterationId := i
		go usersCache.PerformReadOperation(func(userData *UserData) {
			userData.ToApi()
			userData.UpdateData(func(userdata *UserData) {
				userdata.Experience += 10
				userdata.GameLevel = int(userdata.Experience / 100)
			})
		})
		go func() {
			u, _ := usersCache.GetUserData("uid_001")
			u.SetExperience(199)
		}()
		go func() {
			u, _ := usersCache.GetUserData("uid_001")
			u.UpdateData(func(userdata *UserData) {
				userdata.Experience += 10
				userdata.GameLevel = int(userdata.Experience / 100)
			})
		}()
	}

	levelCounts := usersCache.MapReduceUsersWithFilter(excludeJohnFilter, userLevelMapper, levelCountReducer)

	for level, count := range levelCounts.(map[int]int) {
		fmt.Printf("Level %d: %d users\n", level, count)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGINT)
	<-interrupt
	fmt.Println("Stopping server..")
}
