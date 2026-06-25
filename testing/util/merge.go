package util

import (
	silapb "github.com/sila-chain/Sila-Consensus-Core/v7/proto/sila/v1alpha1"
)

// ----------------------------------------------------------------------------
// Bellatrix
// ----------------------------------------------------------------------------

// NewBeaconBlockBellatrix creates a beacon block with minimum marshalable fields.
func NewBeaconBlockBellatrix() *silapb.SignedBeaconBlockBellatrix {
	return HydrateSignedBeaconBlockBellatrix(&silapb.SignedBeaconBlockBellatrix{})
}

// NewBlindedBeaconBlockBellatrix creates a blinded beacon block with minimum marshalable fields.
func NewBlindedBeaconBlockBellatrix() *silapb.SignedBlindedBeaconBlockBellatrix {
	return HydrateSignedBlindedBeaconBlockBellatrix(&silapb.SignedBlindedBeaconBlockBellatrix{})
}

// ----------------------------------------------------------------------------
// Capella
// ----------------------------------------------------------------------------

// NewBeaconBlockCapella creates a beacon block with minimum marshalable fields.
func NewBeaconBlockCapella() *silapb.SignedBeaconBlockCapella {
	return HydrateSignedBeaconBlockCapella(&silapb.SignedBeaconBlockCapella{})
}

// NewBlindedBeaconBlockCapella creates a blinded beacon block with minimum marshalable fields.
func NewBlindedBeaconBlockCapella() *silapb.SignedBlindedBeaconBlockCapella {
	return HydrateSignedBlindedBeaconBlockCapella(&silapb.SignedBlindedBeaconBlockCapella{})
}

// ----------------------------------------------------------------------------
// Deneb
// ----------------------------------------------------------------------------

// NewBeaconBlockDeneb creates a beacon block with minimum marshalable fields.
func NewBeaconBlockDeneb() *silapb.SignedBeaconBlockDeneb {
	return HydrateSignedBeaconBlockDeneb(&silapb.SignedBeaconBlockDeneb{})
}

// NewBeaconBlockContentsDeneb creates a beacon block with minimum marshalable fields.
func NewBeaconBlockContentsDeneb() *silapb.SignedBeaconBlockContentsDeneb {
	return HydrateSignedBeaconBlockContentsDeneb(&silapb.SignedBeaconBlockContentsDeneb{})
}

// NewBlindedBeaconBlockDeneb creates a blinded beacon block with minimum marshalable fields.
func NewBlindedBeaconBlockDeneb() *silapb.SignedBlindedBeaconBlockDeneb {
	return HydrateSignedBlindedBeaconBlockDeneb(&silapb.SignedBlindedBeaconBlockDeneb{})
}

// ----------------------------------------------------------------------------
// Electra
// ----------------------------------------------------------------------------

// NewBeaconBlockElectra creates a beacon block with minimum marshalable fields.
func NewBeaconBlockElectra() *silapb.SignedBeaconBlockElectra {
	return HydrateSignedBeaconBlockElectra(&silapb.SignedBeaconBlockElectra{})
}

// NewBeaconBlockContentsElectra creates a beacon block with minimum marshalable fields.
func NewBeaconBlockContentsElectra() *silapb.SignedBeaconBlockContentsElectra {
	return HydrateSignedBeaconBlockContentsElectra(&silapb.SignedBeaconBlockContentsElectra{})
}

// NewBlindedBeaconBlockElectra creates a blinded beacon block with minimum marshalable fields.
func NewBlindedBeaconBlockElectra() *silapb.SignedBlindedBeaconBlockElectra {
	return HydrateSignedBlindedBeaconBlockElectra(&silapb.SignedBlindedBeaconBlockElectra{})
}

// ----------------------------------------------------------------------------
// Fulu
// ----------------------------------------------------------------------------

// NewBeaconBlockFulu creates a beacon block with minimum marshalable fields.
func NewBeaconBlockFulu() *silapb.SignedBeaconBlockFulu {
	return HydrateSignedBeaconBlockFulu(&silapb.SignedBeaconBlockFulu{})
}

// NewBeaconBlockContentsFulu creates a beacon block with minimum marshalable fields.
func NewBeaconBlockContentsFulu() *silapb.SignedBeaconBlockContentsFulu {
	return HydrateSignedBeaconBlockContentsFulu(&silapb.SignedBeaconBlockContentsFulu{})
}

// NewBlindedBeaconBlockFulu creates a blinded beacon block with minimum marshalable fields.
func NewBlindedBeaconBlockFulu() *silapb.SignedBlindedBeaconBlockFulu {
	return HydrateSignedBlindedBeaconBlockFulu(&silapb.SignedBlindedBeaconBlockFulu{})
}

// ----------------------------------------------------------------------------
// Gloas
// ----------------------------------------------------------------------------

// NewBeaconBlockGloas creates a beacon block with minimum marshalable fields.
func NewBeaconBlockGloas() *silapb.SignedBeaconBlockGloas {
	return HydrateSignedBeaconBlockGloas(&silapb.SignedBeaconBlockGloas{})
}
