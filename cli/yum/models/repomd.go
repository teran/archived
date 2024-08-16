package models

import (
	"encoding/xml"

	"github.com/pkg/errors"
)

type RepoMDDataChecksum struct {
	Text string `xml:",chardata"`
	Type string `xml:"type,attr"`
}

type RepoMDDataOpenChecksum struct {
	Text string `xml:",chardata"`
	Type string `xml:"type,attr"`
}

type RepoMDDataLocation struct {
	Text string `xml:",chardata"`
	Href string `xml:"href,attr"`
}

type RepoMDData struct {
	Text         string                 `xml:",chardata"`
	Type         string                 `xml:"type,attr"`
	Checksum     RepoMDDataChecksum     `xml:"checksum"`
	OpenChecksum RepoMDDataOpenChecksum `xml:"open-checksum"`
	Location     RepoMDDataLocation     `xml:"location"`
	Timestamp    string                 `xml:"timestamp"`
	Size         string                 `xml:"size"`
	OpenSize     string                 `xml:"open-size"`
}

type RepoMD struct {
	XMLName  xml.Name     `xml:"repomd"`
	Text     string       `xml:",chardata"`
	Xmlns    string       `xml:"xmlns,attr"`
	Rpm      string       `xml:"rpm,attr"`
	Revision string       `xml:"revision"`
	Data     []RepoMDData `xml:"data"`
}

func (r *RepoMD) GetPrimary() (RepoMDData, error) {
	for _, md := range r.Data {
		if md.Type == "primary" {
			return md, nil
		}
	}
	return RepoMDData{}, errors.New("no primary MD found")
}
