package data

import (
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "strings"
)

type ClinvarVariant struct {
    HGVS             string // e.g. "p.R273C"
    ClinSignificance string // e.g. "Pathogenic"
    Up               bool   // true for Pathogenic/Likely pathogenic
}

// GetClinvarVariants fetches all missense variants from ClinVar for the given gene symbol.
func GetClinvarVariants(gene string) ([]ClinvarVariant, error) {
    // 1. ESearch to get list of RCV IDs
    esearch := "https://eutils.ncbi.nlm.nih.gov/entrez/eutils/esearch.fcgi"
    params := url.Values{
        "db":    {"clinvar"},
        "term":  {fmt.Sprintf("%s[gene] AND missense[variant type]", gene)},
        "retmax": {"1000"},
        "retmode": {"json"},
    }
    res, err := http.Get(esearch + "?" + params.Encode())
    if err != nil {
        return nil, err
    }
    defer res.Body.Close()

    var e struct {
        ESrchResult struct{ IdList []string } `json:"esearchresult"`
    }
    if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
        return nil, err
    }

    // 2. Batch EFetch to get summaries
    var variants []ClinvarVariant
    for _, id := range e.ESrchResult.IdList {
        efetch := "https://eutils.ncbi.nlm.nih.gov/entrez/eutils/efetch.fcgi"
        params = url.Values{"db": {"clinvar"}, "id": {id}, "retmode": {"json"}}
        resp, err := http.Get(efetch + "?" + params.Encode())
        if err != nil {
            return nil, err
        }
        defer resp.Body.Close()

        var out struct {
            ClinicalAssertionList struct {
                ClinicalAssertion []struct {
                    TraitSet struct {
                        List []struct {
                            XRef struct {
                                ID   string `json:"id"`
                                Type string `json:"db"`
                            } `json:"xref"`
                        } `json:"traitset"`
                    } `json:"traitset"`
                    VariantIdentification struct {
                        HGVS string `json:"hgvs"` // protein HGVS
                    } `json:"variantid"`
                    ClinicalSignificance struct {
                        Description string `json:"description"`
                    } `json:"clinicalsignificance"`
                } `json:"clinicalassertionlist"`
            } `json:"clinicalassertionlist"`
        }
        if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
            return nil, err
        }

        for _, ca := range out.ClinicalAssertionList.ClinicalAssertion {
            hgvs := ca.VariantIdentification.HGVS
            sig := ca.ClinicalSignificance.Description
            up := strings.EqualFold(sig, "Pathogenic") || strings.EqualFold(sig, "Likely pathogenic")
            variants = append(variants, ClinvarVariant{HGVS: hgvs, ClinSignificance: sig, Up: up})
        }
    }
    return variants, nil
}
