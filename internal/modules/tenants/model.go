package tenants

type Bootstrap struct {
	Slug        string
	DisplayName string
	Branding    Branding
}

type Branding struct {
	LogoURL        *string
	PrimaryColor   *string
	SecondaryColor *string
	CustomDomain   *string
}
