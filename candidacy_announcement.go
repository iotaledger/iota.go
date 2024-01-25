package iotago

// CandidacyAnnouncement is a payload which is used to indicate candidacy for committee selection for the next epoch.
type CandidacyAnnouncement struct {
}

func (u *CandidacyAnnouncement) Clone() Payload {
	return &CandidacyAnnouncement{}
}

func (u *CandidacyAnnouncement) PayloadType() PayloadType {
	return PayloadCandidacyAnnouncement
}

func (u *CandidacyAnnouncement) Size() int {
	// PayloadType
	return 1
}

func (u *CandidacyAnnouncement) WorkScore(workScoreParameters *WorkScoreParameters) (WorkScore, error) {
	// we account for the network traffic only on "Payload" level
	return workScoreParameters.DataByte.Multiply(u.Size())
}
