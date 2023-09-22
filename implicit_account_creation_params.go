package iotago

type ImplicitAccountCreationParameters struct {
	// OnboardingReferenceManaCost is the value of the reference Mana cost used to prefund a BasicOutput on an ImplicitAccountCreationAddress.
	OnboardingReferenceManaCost Mana `serix:"0,mapKey=onboardingReferenceManaCost"`
}

func (c *ImplicitAccountCreationParameters) Equals(other ImplicitAccountCreationParameters) bool {
	return c.OnboardingReferenceManaCost == other.OnboardingReferenceManaCost
}
