package rcsv

import (
	"fmt"
	"rentroll/rlib"
	"strings"
)

// CSV FIELDS FOR THIS MODULE
// 0    1          2
// Name,RARequired,Description
// Rent,0,"Rent: the recurring amount due under an Occupancy Agreement.  While most residential leases are one year or less, commecial leases may go on decades.  In those

// CreateAssessmentType reads an assessment type string array and creates a database record for the assessment type
func CreateAssessmentType(sa []string, lineno int) {
	funcname := "CreateAssessmentType"
	des := strings.TrimSpace(sa[0])
	if strings.ToLower(des) == "name" {
		return // this is just the column heading
	}

	// fmt.Printf("line %d, sa = %#v\n", lineno, sa)
	if len(sa) < 3 {
		fmt.Printf("%s: line %d - found %d values, there must be at least 3\n", funcname, lineno, len(sa))
		return
	}

	//-------------------------------------------------------------------
	// Check to see if this assessment type is already in the database
	//-------------------------------------------------------------------
	if len(des) > 0 {
		a1, _ := rlib.GetAssessmentTypeByName(des)
		if len(a1.Name) > 0 {
			rlib.Ulog("%s: rlib.AssessmentType named %s already exists\n", funcname, des)
			return
		}
	}

	var a rlib.AssessmentType
	a.Name = strings.TrimSpace(sa[0])
	if len(a.Name) == 0 {
		rlib.Ulog("%s: line %d - Name cannot be empty\n", funcname, lineno)
		return
	}

	//-------------------------------------------------------------------
	// RARequired
	//-------------------------------------------------------------------
	a.RARequired, _ = rlib.IntFromString(sa[1], "RARequired value is invalid")
	if a.RARequired < rlib.RARQDINRANGE || a.RARequired > rlib.RARQDANY {
		fmt.Printf("%s: line %d - RARequired must be in the range %d to %d.  Found: %s\n", funcname, lineno, rlib.RARQDINRANGE, rlib.RARQDANY, sa[1])
		return
	}

	a.Description = sa[2]
	rlib.Errlog(rlib.InsertAssessmentType(&a))
}

// LoadAssessmentTypesCSV loads a csv file with assessment types and processes each one
func LoadAssessmentTypesCSV(fname string) {
	t := rlib.LoadCSV(fname)
	for i := 0; i < len(t); i++ {
		CreateAssessmentType(t[i], i+1)
	}
}