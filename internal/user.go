package internal

import (
	"fmt"
	"io/ioutil"
	"os"
)

// User is the data structure to represent a single user.
type User struct {
	YammerUserResponse
	mugShot []byte
	mugFile *os.File
	Groups  *YammerGroupResponse
}

// Users is the data structure to represent all users.
type Users struct {
	client *Client
	cache  map[int64]*User
	tmpdir string
}

// NewUsers returns a new Users object.
func NewUsers(client *Client, tmpdir string) *Users {

	return &Users{
		client: client,
		tmpdir: tmpdir,
		cache:  make(map[int64]*User),
	}
}

// GetUser returns the user by id.
func (users *Users) GetUser(uid int64) (*User, error) {

	// get user from cache
	if user, ok := users.cache[uid]; ok {
		return user, nil
	}

	// construct path (current by default, for a particular group if uid is !-1)
	pathUser := "users/current.json"
	if uid != -1 {
		pathUser = fmt.Sprintf("users/%d.json", uid)
	}

	// construct request
	reqUser, errReq := users.client.newRequest("GET", pathUser, nil, nil)
	if errReq != nil {
		return nil, fmt.Errorf("failed to construct user request for user %d: %v", uid, errReq)
	}

	// do request and parse response
	var yur YammerUserResponse
	_, errUserDo := users.client.do(reqUser, &yur)
	if errUserDo != nil {
		return nil, fmt.Errorf("failed to do user request for user %d: %v", uid, errUserDo)
	}

	// query yammer mug shot
	mug, errMug := users.client.GetImage(yur.MugshotURL)
	if errMug != nil {
		return nil, fmt.Errorf("failed to get mug shot for user: %d: %v", uid, errMug)
	}

	// if user is current user query groups
	var groups *YammerGroupResponse
	if uid == -1 {

		// construct path
		pathGroups := fmt.Sprintf("groups/for_user/%d.json", yur.ID)

		// construct request
		reqGrp, errGrp := users.client.newRequest("GET", pathGroups, nil, nil)
		if errGrp != nil {
			return nil, fmt.Errorf("failed to construct groups request for user %d: %v", uid, errGrp)
		}

		// do request and parse response
		var ygr YammerGroupResponse
		_, errGrpDo := users.client.do(reqGrp, &ygr)
		if errGrpDo != nil {
			return nil, fmt.Errorf("failed to do groups request for user %d: %v", uid, errGrpDo)
		}

		// append the -1-group for private messages
		privateGroup := YammerGroup{
			ID:       -1,
			FullName: "Private",
		}
		ygr = append(ygr, privateGroup)

		groups = &ygr
	}

	// construct our user
	user := User{yur, mug, nil, groups}

	// update cache
	users.cache[uid] = &user

	// return user
	return &user, nil
}

func (users *Users) GetMugFile(user *User) (*os.File, error) {

	if user.mugFile != nil && FileExists(user.mugFile.Name()) {
		return user.mugFile, nil
	}
	file, errDump := users.DumpMugToFile(user)
	if errDump != nil {
		return nil, fmt.Errorf("failed to dump mug: %v", errDump)
	}
	user.mugFile = file
	return file, nil

}

func (users *Users) DumpMugToFile(user *User) (*os.File, error) {

	// if we don't have a mug
	if user.mugShot == nil || len(user.mugShot) < 1 {
		return nil, fmt.Errorf("nothing to dump")
	}

	//Create a temp file
	file, errTmp := ioutil.TempFile(users.tmpdir, fmt.Sprintf("goyammer_%d_*.jpg", user.ID))
	if errTmp != nil {
		return nil, fmt.Errorf("couldn't create temp file: %v", errTmp)
	}
	defer func() {
		_ = file.Close()
	}()

	//Write the bytes to the file
	_, errWrite := file.Write(user.mugShot)
	if errWrite != nil {
		return nil, fmt.Errorf("couldn't write mug to temp file: %v", errWrite)
	}

	return file, nil
}
