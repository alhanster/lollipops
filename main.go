package main

import (
    "flag"
    "fmt"
    "os"
    "strings"

    "github.com/joiningdata/lollipops/data"
    "github.com/joiningdata/lollipops/drawing"
)

func main() {
    gene := flag.String("gene", "", "HGNC gene symbol to fetch missense variants from ClinVar")
    // you can carry over any of your existing flags (-o, -w, -baseline-colors, etc.)
    flag.Parse()

    if *gene == "" {
        fmt.Fprintln(os.Stderr, "error: -gene flag is required")
        flag.Usage()
        os.Exit(1)
    }

    // 1) Lookup UniProt ID & domains as before
    uniprotID, err := data.LookupUniProtID(*gene)
    if err != nil {
        panic(err)
    }
    domains, err := data.GetGraphicData(uniprotID)
    if err != nil {
        panic(err)
    }

    // 2) Fetch ClinVar missense variants for this gene
    clinvars, err := data.GetClinVarMissense(*gene)
    if err != nil {
        panic(err)
    }

    // 3) Build lollipop mutation tags, flipping non‐pathogenic ones downward
    var muts []string
    for _, cv := range clinvars {
        tag := convertHGVS(cv.ProteinChange)
        // “Pathogenic” & “Likely pathogenic” → default (up)
        if strings.EqualFold(cv.Classification, "pathogenic") ||
           strings.EqualFold(cv.Classification, "likely pathogenic") {
            muts = append(muts, tag)
        } else {
            // add a tiny negative count to flip orientation
            muts = append(muts, fmt.Sprintf("%s@-1", tag))
        }
    }

    // 4) Draw
    settings := drawing.DefaultSettings()
    // copy any flags you like into settings here…
    drawing.DrawSVGWithSettings(os.Stdout, muts, domains, settings)
}

// convertHGVS turns “p.Asn754Ser” → “N754S”
func convertHGVS(hgvs string) string {
    hgvs = strings.TrimPrefix(hgvs, "p.")
    threeToOne := map[string]string{
        "Ala":"A","Cys":"C","Asp":"D","Glu":"E","Phe":"F",
        "Gly":"G","His":"H","Ile":"I","Lys":"K","Leu":"L",
        "Met":"M","Asn":"N","Pro":"P","Gln":"Q","Arg":"R",
        "Ser":"S","Thr":"T","Val":"V","Trp":"W","Tyr":"Y",
    }
    orig := hgvs[:3]
    pos := hgvs[3: len(hgvs)-3]
    nw  := hgvs[len(hgvs)-3:]
    return threeToOne[orig] + pos + threeToOne[nw]
}
