package apt

import (
	debian "pault.ag/go/debian/control"
)

type RepositoryRelease struct {
	Origin                      string                  `control:"Origin"`
	Label                       string                  `control:"Label"`
	Suite                       string                  `control:"Suite"`
	Version                     string                  `control:"Version"`
	Codename                    string                  `control:"Codename"`
	Changelogs                  string                  `control:"Changelogs"`
	Date                        string                  `control:"Date"`
	AcquireByHash               bool                    `control:"Acquire-By-Hash"`
	NoSupportForArchitectureAll string                  `control:"No-Support-for-Architecture-all"`
	Architectures               []string                `control:"Architectures" delim:" "`
	Components                  []string                `control:"Components" delim:" "`
	Description                 string                  `control:"Description"`
	MD5Sum                      []debian.MD5FileHash    `control:"MD5Sum" delim:"\n" strip:"\n\r\t "`
	SHA256Sum                   []debian.SHA256FileHash `control:"SHA256" delim:"\n" strip:"\n\r\t "`
}

type ComponentRelease struct {
	Archive       string `control:"Archive"`
	Origin        string `control:"Origin"`
	Label         string `control:"Label"`
	Version       string `control:"Version"`
	AcquireByHash string `control:"Acquire-By-Hash"`
	Component     string `control:"Component"`
	Architecture  string `control:"Architecture"`
}

type Package struct {
	Version        string   `control:"Version"`
	InstalledSize  int      `control:"Installed-Size"`
	Maintainer     string   `control:"Maintainer"`
	Architecture   string   `control:"Architecture"`
	Depends        []string `control:"Depends" delim:"," strip:"\n\r\t "`
	PreDepends     string   `control:"Pre-Depends" delim:"," strip:"\n\r\t "`
	Description    string   `control:"Description"`
	Homepage       string   `control:"Homepage"`
	DescriptionMD5 string   `control:"Description-md5"`
	Tag            []string `control:"Tag" delim:"," strip:"\n\r\t "`
	Section        string   `control:"Section"`
	Priority       string   `control:"Priority"`
	Filename       string   `control:"Filename"`
	Size           int      `control:"Size"`
	MD5Sum         string   `control:"MD5sum"`
	SHA256Sum      string   `control:"SHA256"`
}

type Packages []Package
