# Auth

## GetAccountID

We offer a convenient way to extract the AccountID field from an incoming authorization token.
For this purpose `auth.GetAccountID(ctx, nil)` function can be used:
```
func (s *contactsServer) Read(ctx context.Context, req *ReadRequest) (*ReadResponse, error) {
	input := req.GetContact()

	accountID, err := auth.GetAccountID(ctx, nil)
	if err == nil {
		input.AccountId = accountID
	} else if input.GetAccountId() == "" {
		return nil, err
	}

	c, err := DefaultReadContact(ctx, input, s.db)
	if err != nil {
		return nil, err
	}
	return &ReadResponse{Contact: c}, nil
}
```

When bootstrapping a gRPC server, add middleware that will extract the account_id token from the request context and set it in the request struct. The middleware will have to navigate the request struct via reflection, in the case that the account_id field is nested within the request (like if it's in a request wrapper as per our example above)