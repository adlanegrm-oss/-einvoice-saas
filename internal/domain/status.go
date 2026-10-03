package domain

type ComplianceStatus string

const (
	ComplianceNotValidated ComplianceStatus = "NOT_VALIDATED"
	ComplianceValid        ComplianceStatus = "VALID"
	ComplianceInvalid      ComplianceStatus = "INVALID"
)

type TransmissionStatus string

const (
	TransmissionNotSubmitted TransmissionStatus = "NOT_SUBMITTED"
	TransmissionPending      TransmissionStatus = "PENDING"
	TransmissionSubmitted    TransmissionStatus = "SUBMITTED"
	TransmissionAccepted     TransmissionStatus = "ACCEPTED"
	TransmissionDelivered    TransmissionStatus = "DELIVERED"
	TransmissionRejected     TransmissionStatus = "REJECTED"
	TransmissionFailed       TransmissionStatus = "FAILED"
)

type PaymentStatus string

const (
	PaymentNotDue        PaymentStatus = "NOT_DUE"
	PaymentDue           PaymentStatus = "DUE"
	PaymentPartiallyPaid PaymentStatus = "PARTIALLY_PAID"
	PaymentPaid          PaymentStatus = "PAID"
)

type ExternalState string

const (
	ExtDraft      ExternalState = "DRAFT"
	ExtValidating ExternalState = "VALIDATING"
	ExtReady      ExternalState = "READY"
	ExtSubmitting ExternalState = "SUBMITTING"
	ExtSubmitted  ExternalState = "SUBMITTED"
	ExtDelivered  ExternalState = "DELIVERED"
	ExtRejected   ExternalState = "REJECTED"
	ExtFailed     ExternalState = "FAILED"
	ExtPaid       ExternalState = "PAID"
)

func ComputeExternalState(comp ComplianceStatus, trans TransmissionStatus, pay PaymentStatus) ExternalState {
	if pay == PaymentPaid {
		return ExtPaid
	}
	if trans == TransmissionDelivered {
		return ExtDelivered
	}
	if trans == TransmissionRejected {
		return ExtRejected
	}
	if trans == TransmissionFailed {
		return ExtFailed
	}
	if trans == TransmissionSubmitted || trans == TransmissionAccepted {
		return ExtSubmitted
	}
	if trans == TransmissionPending {
		return ExtSubmitting
	}
	if comp == ComplianceInvalid {
		return ExtRejected
	}
	if comp == ComplianceValid {
		return ExtReady
	}
	return ExtDraft
}
