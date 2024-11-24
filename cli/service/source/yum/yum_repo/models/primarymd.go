package models

import "encoding/xml"

type PrimaryMDPackageVersion struct {
	Text  string `xml:",chardata"`
	Epoch string `xml:"epoch,attr"`
	Ver   string `xml:"ver,attr"`
	Rel   string `xml:"rel,attr"`
}

type PrimaryMDPackageChecksum struct {
	Text  string `xml:",chardata"`
	Type  string `xml:"type,attr"`
	PkgID string `xml:"pkgid,attr"`
}

type PrimaryMDPackageTime struct {
	Text  string `xml:",chardata"`
	File  string `xml:"file,attr"`
	Build string `xml:"build,attr"`
}

type PrimaryMDPackageSize struct {
	Text      string `xml:",chardata"`
	Package   uint64 `xml:"package,attr"`
	Installed string `xml:"installed,attr"`
	Archive   string `xml:"archive,attr"`
}

type PrimaryMDPackageLocation struct {
	Text string `xml:",chardata"`
	Href string `xml:"href,attr"`
}

type PrimaryMDPackageFormatFile struct {
	Text string `xml:",chardata"`
	Type string `xml:"type,attr"`
}

type PrimaryMDPackageFormat struct {
	Text      string                       `xml:",chardata"`
	License   string                       `xml:"license"`
	Vendor    string                       `xml:"vendor"`
	Group     string                       `xml:"group"`
	BuildHost string                       `xml:"buildhost"`
	SourceRPM string                       `xml:"sourcerpm"`
	File      []PrimaryMDPackageFormatFile `xml:"file"`
}

type PrimaryMDPackage struct {
	Text        string                   `xml:",chardata"`
	Type        string                   `xml:"type,attr"`
	Name        string                   `xml:"name"`
	Arch        string                   `xml:"arch"`
	Version     PrimaryMDPackageVersion  `xml:"version"`
	Checksum    PrimaryMDPackageChecksum `xml:"checksum"`
	Summary     string                   `xml:"summary"`
	Description string                   `xml:"description"`
	Packager    string                   `xml:"packager"`
	URL         string                   `xml:"url"`
	Time        PrimaryMDPackageTime     `xml:"time"`
	Size        PrimaryMDPackageSize     `xml:"size"`
	Location    PrimaryMDPackageLocation `xml:"location"`
	Format      PrimaryMDPackageFormat   `xml:"format"`
}

type PrimaryMD struct {
	XMLName  xml.Name           `xml:"metadata"`
	Text     string             `xml:",chardata"`
	Xmlns    string             `xml:"xmlns,attr"`
	Rpm      string             `xml:"rpm,attr"`
	Packages string             `xml:"packages,attr"`
	Package  []PrimaryMDPackage `xml:"package"`
}
