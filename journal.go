package main

import (
	"fmt"
	"math"
	"rentroll/rlib"
	"strconv"
	"strings"
	"time"
)

func roundToCent(x float32) float32 {
	return float32(int(x*float32(100)+float32(0.5))) / float32(100)
}

func sumAllocations(m *[]acctRule) (float32, float32) {
	sum := float32(0.0)
	debits := float32(0.0)
	for i := 0; i < len(*m); i++ {
		if (*m)[i].Action == "c" {
			sum -= (*m)[i].Amount
		} else {
			sum += (*m)[i].Amount
			debits += (*m)[i].Amount
		}
	}
	return sum, debits
}

// journalAssessment processes the assessment, creates a journal entry, and returns its id
func journalAssessment(d time.Time, a *Assessment, d1, d2 *time.Time) error {
	//-------------------------------------------------------------------
	// over what range of time does this rental apply between d1 & d2
	//-------------------------------------------------------------------
	ra, _ := GetRentalAgreement(a.RAID)
	start := *d1
	if ra.RentalStart.After(start) {
		start = ra.RentalStart
	}
	stop := ra.RentalStop.Add(24 * 60 * time.Minute)
	if stop.After(*d2) {
		stop = *d2
	}
	//-------------------------------------------------------------------------------------------
	// this code needs to be generalized based on the recurrence period and the proration period
	//-------------------------------------------------------------------------------------------
	assessmentDuration := int(d2.Sub(*d1).Hours() / 24)
	rentDuration := int(stop.Sub(start).Hours() / 24)
	pf := float32(1.0)
	if rentDuration != assessmentDuration && a.ProrationMethod > 0 {
		pf = float32(rentDuration) / float32(assessmentDuration)
	} else {
		rentDuration = assessmentDuration
	}

	var j Journal
	j.BID = a.BID
	j.Dt = d
	j.Type = JNLTYPEASMT
	j.ID = a.ASMID
	j.RAID = a.RAID

	m := parseAcctRule(a.AcctRule, pf) // a rule such as "d 11001 1000.0, c 40001 1100.0, d 41004 100.00"
	_, j.Amount = sumAllocations(&m)

	//-------------------------------------------------------------------------------------------
	// In the event that we need to prorate, pull together the pieces and determine the
	// fractional amounts so that all the entries can net to 0.00.  Essentially, this means
	// handling the $0.01 off problem when dealing with fractional numbers.  The way we'll
	// handle this is to apply the extra cent to the largest number
	//-------------------------------------------------------------------------------------------
	if pf < 1.0 {
		sum := float32(0.0)
		debits := float32(0)
		k := 0 // index of the largest number
		for i := 0; i < len(m); i++ {
			m[i].Amount = roundToCent(m[i].Amount)
			if m[i].Amount > m[k].Amount {
				k = i
			}
			if m[i].Action == "c" {
				sum -= m[i].Amount
			} else {
				sum += m[i].Amount
				debits += m[i].Amount
			}
		}
		if sum != float32(0) {
			m[k].Amount += sum // first try adding the penny
			x, xd := sumAllocations(&m)
			j.Amount = xd
			if x != float32(0) { // if that doesn't work...
				m[k].Amount -= sum + sum // subtract the penny
				y, yd := sumAllocations(&m)
				j.Amount = yd
				// if there's some strange number that causes issues, use the one closest to 0
				if math.Abs(float64(y)) > math.Abs(float64(x)) { // if y is farther from 0 than x, go back to the value for x
					m[k].Amount += sum + sum
					j.Amount = xd
				}
			}
		}
	}

	jid, err := InsertJournalEntry(&j)
	if err != nil {
		ulog("error inserting journal entry: %v\n", err)
	} else {
		//now rewrite the AcctRule...
		s := ""
		for i := 0; i < len(m); i++ {
			s += fmt.Sprintf("%s %s %.2f", m[i].Action, m[i].Account, roundToCent(m[i].Amount))
			if i+1 < len(m) {
				s += ", "
			}
		}
		if jid > 0 {
			var ja JournalAllocation
			ja.JID = jid
			ja.ASMID = a.ASMID
			ja.Amount = j.Amount
			ja.AcctRule = s
			InsertJournalAllocationEntry(&ja)
		}
	}

	return err
}

func parseAcctRule(rule string, pf float32) []acctRule {
	var m []acctRule
	if len(rule) > 0 {
		sa := strings.Split(rule, ",")
		for k := 0; k < len(sa); k++ {
			var r acctRule
			t := strings.TrimSpace(sa[k])
			ta := strings.Split(t, " ")
			r.Action = strings.ToLower(strings.TrimSpace(ta[0]))
			r.Account = strings.TrimSpace(ta[1])
			f, _ := strconv.ParseFloat(strings.TrimSpace(ta[2]), 64)
			r.Amount = float32(f) * pf
			m = append(m, r)
		}
	}
	return m
}

// RemoveJournalEntries clears out the records in the supplied range provided the range is not closed by a journalmarker
func RemoveJournalEntries(xprop *XBusiness, d1, d2 *time.Time) error {
	// Remove the journal entries and the journalallocation entries
	rows, err := App.prepstmt.getAllJournalsInRange.Query(xprop.P.BID, d1, d2)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var j Journal
		rlib.Errcheck(rows.Scan(&j.JID, &j.BID, &j.RAID, &j.Dt, &j.Amount, &j.Type, &j.ID))
		deleteJournalAllocations(j.JID)
		deleteJournalEntry(j.JID)
	}

	// only delete the marker if it is in this time range and if it is not the origin marker
	jm := GetLastJournalMarker()
	if jm.State == MARKERSTATEOPEN {
		deleteJournalMarker(jm.JMID)
	}

	return err
}

// GenerateJournalRecords creates journal records for assessments and receipts over the supplied time range.
func GenerateJournalRecords(xprop *XBusiness, d1, d2 *time.Time) {
	err := RemoveJournalEntries(xprop, d1, d2)
	if err != nil {
		ulog("Could not remove existin Journal Entries from %s to %s\n", d1.Format(RRDATEFMT), d2.Format(RRDATEFMT))
		return
	}

	//===========================================================
	//  PROCESS ASSESSMSENTS
	//===========================================================
	rows, err := App.prepstmt.getAllAssessmentsByBusiness.Query(xprop.P.BID, d2, d1)
	rlib.Errcheck(err)
	defer rows.Close()
	for rows.Next() {
		var a Assessment
		ap := &a
		rlib.Errcheck(rows.Scan(&a.ASMID, &a.BID, &a.RID, &a.UNITID, &a.ASMTID, &a.RAID, &a.Amount, &a.Start, &a.Stop, &a.Frequency, &a.ProrationMethod, &a.AcctRule, &a.LastModTime, &a.LastModBy))
		if a.Frequency >= rlib.RECURSECONDLY && a.Frequency <= rlib.RECURHOURLY {
			// TBD
			fmt.Printf("Unhandled assessment recurrence type: %d\n", a.Frequency)
		} else {
			dl := ap.GetRecurrences(d1, d2)
			// fmt.Printf("type = %d, %s - %s    len(dl) = %d\n", a.ASMTID, a.Start.Format(RRDATEFMT), a.Stop.Format(RRDATEFMT), len(dl))
			for i := 0; i < len(dl); i++ {
				journalAssessment(dl[i], &a, d1, d2)
			}
		}
	}
	rlib.Errcheck(rows.Err())

	//===========================================================
	//  PROCESS RECEIPTS
	//===========================================================
	r := GetReceipts(xprop.P.BID, d1, d2)
	for i := 0; i < len(r); i++ {
		rntagr, _ := GetRentalAgreement(r[i].RAID)
		var j Journal
		j.BID = rntagr.BID
		j.Amount = r[i].Amount
		j.Dt = r[i].Dt
		j.Type = JNLTYPERCPT
		j.ID = r[i].RCPTID
		j.RAID = r[i].RAID
		jid, err := InsertJournalEntry(&j)
		if err != nil {
			ulog("Error inserting journal entry: %v\n", err)
		}
		if jid > 0 {
			// now add the journal allocation records...
			for j := 0; j < len(r[i].RA); j++ {
				var ja JournalAllocation
				ja.JID = jid
				ja.Amount = r[i].RA[j].Amount
				ja.ASMID = r[i].RA[j].ASMID
				ja.AcctRule = r[i].RA[j].AcctRule
				InsertJournalAllocationEntry(&ja)
			}
		}
	}

	//===========================================================
	//  ADD JOURNAL MARKER
	//===========================================================
	var jm JournalMarker
	jm.BID = xprop.P.BID
	jm.State = MARKERSTATEOPEN
	jm.DtStart = *d1
	jm.DtStop = (*d2).AddDate(0, 0, -1)
	InsertJournalMarker(&jm)
}
