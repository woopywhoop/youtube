package main

import (
	"encoding/json"
	"sync"
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

func LoadUsersDataFromDB(usersCache *UsersCache) error {
	// Mock for actual implementation
	usersCache.AddUserData(
		NewUserData("uid_001", "king", 1, 100),
		NewUserData("uid_002", "queen", 1, 100),
		NewUserData("uid_003", "soldier", 1, 100),
	)
	return nil
}

func main() {
	usersCache := NewUsersCache()
	_ = LoadUsersDataFromDB(usersCache)
}
