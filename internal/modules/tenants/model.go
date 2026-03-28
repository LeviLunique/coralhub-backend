package tenants

type Bootstrap struct {
	Slug        string
	DisplayName string
	Branding    Branding
}

type Context struct {
	ID          string
	Slug        string
	DisplayName string
}

type Branding struct {
	LogoURL        *string
	PrimaryColor   *string
	SecondaryColor *string
	CustomDomain   *string
}
