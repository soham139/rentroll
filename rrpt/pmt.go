package rrpt

import (
	"fmt"
	"gotable"
	"rentroll/rlib"
	"sort"
)

// RRreportPaymentTypesTable generates a table object of all rlib.PaymentType for BID
func RRreportPaymentTypesTable(ri *ReporterInfo) gotable.Table {
	funcname := "RRreportPaymentTypesTable"

	// table init
	tbl := getRRTable()

	tbl.AddColumn("PMTID", 11, gotable.CELLSTRING, gotable.COLJUSTIFYLEFT)
	tbl.AddColumn("BID", 10, gotable.CELLSTRING, gotable.COLJUSTIFYLEFT)
	tbl.AddColumn("Name", 10, gotable.CELLSTRING, gotable.COLJUSTIFYLEFT)
	tbl.AddColumn("Description", 30, gotable.CELLSTRING, gotable.COLJUSTIFYLEFT)

	// set table title, sections
	err := TableReportHeaderBlock(&tbl, "Payment Types", funcname, ri)
	if err != nil {
		rlib.LogAndPrintError(funcname, err)
		return tbl
	}

	m := rlib.GetPaymentTypesByBusiness(ri.Bid)
	var keys []int
	for k := range m {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)

	for _, k := range keys {
		i := int64(k)
		v := m[i]
		tbl.AddRow()
		tbl.Puts(-1, 0, v.IDtoString())
		tbl.Puts(-1, 1, fmt.Sprintf("B%08d", v.BID))
		tbl.Puts(-1, 2, v.Name)
		tbl.Puts(-1, 3, v.Description)
	}
	tbl.TightenColumns()
	return tbl
}

// RRreportPaymentTypes generates a report of all rlib.GLAccount accounts
func RRreportPaymentTypes(ri *ReporterInfo) string {
	tbl := RRreportPaymentTypesTable(ri)
	return ReportToString(&tbl, ri)
}
