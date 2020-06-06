package aws

/*
INFO:	this module contains logic for managing aws sessions
USAGE:	get a session, which you can use to get aws clients for various services.
		The session should be recycled instead of making a new one for every request
*/

import (
	"github.com/aws/aws-sdk-go/aws/session"
)

/*GetSession returns a new Session*/
func GetSession() (*session.Session, error) {
	return session.NewSession(nil)
}
