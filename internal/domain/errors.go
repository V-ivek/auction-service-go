package domain

import "errors"

// Domain errors
var (
	// User errors
	ErrInvalidUsername = errors.New("invalid username")
	ErrInvalidEmail    = errors.New("invalid email")
	ErrUserNotFound    = errors.New("user not found")

	// Listing errors
	ErrInvalidTitle              = errors.New("invalid title")
	ErrInvalidStartingBid        = errors.New("invalid starting bid")
	ErrReserveBelowStarting      = errors.New("reserve price cannot be below starting bid")
	ErrInvalidOwner              = errors.New("invalid owner")
	ErrInvalidStatusTransition   = errors.New("invalid status transition")
	ErrCannotUpdateActiveListing = errors.New("cannot update active listing")
	ErrListingNotFound           = errors.New("listing not found")

	// Auction errors
	ErrInvalidListingID      = errors.New("invalid listing ID")
	ErrInvalidDuration       = errors.New("invalid duration")
	ErrAuctionNotFound       = errors.New("auction not found")
	ErrAuctionNotActive      = errors.New("auction is not active")
	ErrAuctionAlreadyStarted = errors.New("auction already started")
	ErrCannotExtendAuction   = errors.New("cannot extend auction")

	// Bid errors
	ErrInvalidAuctionID = errors.New("invalid auction ID")
	ErrInvalidBidder    = errors.New("invalid bidder")
	ErrInvalidBidAmount = errors.New("invalid bid amount")
	ErrBidTooLow        = errors.New("bid amount is too low")
	ErrBidNotFound      = errors.New("bid not found")
	ErrCannotBidOnOwnItem = errors.New("cannot bid on own item")

	// Subscription errors
	ErrInvalidSubscription   = errors.New("invalid subscription")
	ErrSubscriptionNotFound  = errors.New("subscription not found")
	ErrAlreadySubscribed     = errors.New("already subscribed")

	// General errors
	ErrInternalError        = errors.New("internal error")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrForbidden            = errors.New("forbidden")
	ErrNotImplemented       = errors.New("not implemented")
)