package data

import (
    "encoding/xml"
    "fmt"
    "net/http"
    "net/url"
    "strings"
)

// ClinVarVariant holds a protein‐level change (HGVS) and its classification.
type ClinVarVariant struct {
    ProteinChange  string
    Classification string
}

// GetClinVarMissense retrieves all human missense variants for `gene`
// from ClinVar via the Entrez E-utilities (esearch + efetch) :contentReference[oaicite:0]{index=0}.
func GetClinVarMissense(gene string) ([]ClinVarVariant, error) {
    // 1) esearch to get ClinVar IDs
    esearchURL := "https://eutils.ncbi.nlm.nih.gov/entrez/eutils/esearch.fcgi"
    params := url.Values{
        "db":      {"clinvar"},
        "term":    {fmt.Sprintf("%s[gene] AND missense AND human[orgn]", gene)},
        "retmax":  {"10000"},
        "retmode": {"xml"},
    }
    resp, err := http.Get(esearchURL + "?" + params.Encode())
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var es ESearchResult
    if err := xml.NewDecoder(resp.Body).Decode(&es); err != nil {
        return nil, err
    }
    if len(es.IDList) == 0 {
        return nil, nil
    }

    // 2) efetch to retrieve variant details
    efetchURL := "https://eutils.ncbi.nlm.nih.gov/entrez/eutils/efetch.fcgi"
    params2 := url.Values{
        "db":      {"clinvar"},
        "retmode": {"xml"},
        "id":      {strings.Join(es.IDList, ",")},
    }
    resp2, err := http.Get(efetchURL + "?" + params2.Encode())
    if err != nil {
        return nil, err
    }
    defer resp2.Body.Close()

    var records ClinVarRecords
    if err := xml.NewDecoder(resp2.Body).Decode(&records); err != nil {
        return nil, err
    }

    // 3) Extract only those records with a protein‐level HGVS change
    var out []ClinVarVariant
    for _, r := range records.Variants {
        if r.HGVSProtein == "" {
            continue
        }
        out = append(out, ClinVarVariant{
            ProteinChange:  r.HGVSProtein,
            Classification: r.ClinicalSignificance,
        })
    }
    return out, nil
}

// ESearchResult represents the output of esearch.fcgi
type ESearchResult struct {
    XMLName xml.Name `xml:"eSearchResult"`
    IDList  []string `xml:"IdList>Id"`
}

// ClinVarRecords represents the efetch output
type ClinVarRecords struct {
    XMLName  xml.Name        `xml:"ClinVarSet"`
    Variants []ClinVarRecord `xml:"ReferenceClinVarAssertion"`
}

// ClinVarRecord extracts the protein change and clinical significance
type ClinVarRecord struct {
    // Adjust the XPath below to match exactly where the HGVS(p.) is in the XML:
    HGVSProtein         string `xml:"ClinVarAssertion>TraitSet>Trait>Name>ElementValue"`
    ClinicalSignificance string `xml:"ClinicalSignificance>Description"`
}
