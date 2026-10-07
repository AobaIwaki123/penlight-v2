package model

// UpdateMemberPenlightRequest represents the payload to update a member's penlight colors.
type UpdateMemberPenlightRequest struct {
	LeftColorID  ID   `json:"left_color_id"`
	RightColorID ID   `json:"right_color_id"`
	Ordered      bool `json:"ordered"`
}

// UpdateMemberStatusRequest represents the payload to update a member's activity status and/or generation.
type UpdateMemberStatusRequest struct {
	Status     *MemberStatus `json:"status,omitempty"`
	Generation *int          `json:"generation,omitempty"`
}

// SetPrimaryMemberImageRequest represents the payload to set a member's primary/default image.
type SetPrimaryMemberImageRequest struct {
	ImageID ID `json:"image_id"`
}

// UpdateMemberImagePhotoTypeRequest represents the payload to update a member image's costume/photo category.
type UpdateMemberImagePhotoTypeRequest struct {
	PhotoTypeID ID `json:"photo_type_id"`
}
