package cache

// KeyAccount returns a cache key to be used for account values.
func KeyAccount(id string) string {
	return "Account::" + id
}

// KeyAccountName returns a cache key to be used for account by name values.
func KeyAccountName(name string) string {
	return "Account::Name::" + name
}

// KeyUser returns a cache key to be used for user values.
func KeyUser(id string) string {
	return "User::" + id
}

// KeyUserDetails returns a cache key to be used for user details values.
func KeyUserDetails(id string) string {
	return "User::Details::" + id
}

// KeyAuthToken returns a cache key to be used for authentication token values.
func KeyAuthToken(token string) string {
	return "Token::Auth::" + token
}

// KeyToken returns a cache key to be used for token values.
func KeyToken(token string) string {
	return "Token::" + token
}

// KeyResource returns a cache key to be used for resource values.
func KeyResource(id string) string {
	return "Resource::" + id
}
