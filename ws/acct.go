package ws

import (
	"encoding/json"
	"fmt"
	"net/http"
	"rentroll/bizlogic"
	"rentroll/rlib"
	"sort"
	// "strconv"
	"strings"
	"time"
)

// ListedAccount is struct to list down individual ledger Account record
type ListedAccount struct {
	LID  int64  `json:"id"`   // Ledger account ID
	Name string `json:"text"` // Ledger account name
}

// AccountListResponse is the response to list down all ledger accounts
type AccountListResponse struct {
	Status  string          `json:"status"`
	Total   int64           `json:"total"`
	Records []ListedAccount `json:"records"`
}

// // w2uiChild struct used to build subgrid
// type w2uiChild struct {
// 	Children []GLAccount `json:"children"`
// }

// GLAccount describes the static (or mostly static) attributes of a Ledger
type GLAccount struct {
	Recid       int               `json:"recid"` // this is for the grid widget
	LID         int64             // unique id for this GLAccount
	PLID        int64             // unique id of Parent, 0 if no parent
	BID         int64             // Business unit associated with this GLAccount
	RAID        int64             // associated rental agreement, this field is only used when Type = 1
	TCID        int64             // associated payor, this field is only used when Type = 1
	GLNumber    string            // acct system name
	Status      int64             // Whether a GL Account is currently unknown=0, inactive=1, active=2
	Name        string            // descriptive name for the GLAccount
	AcctType    string            // QB Acct Type: Income, Expense, Fixed Asset, Bank, Loan, Credit Card, Equity, Accounts Receivable, Other Current Asset, Other Asset, Accounts Payable, Other Current Liability, Cost of Goods Sold, Other Income, Other Expense
	AllowPost   int64             // 0 = no posting, 1 = posting is allowed
	RARequired  int64             // 0 = during rental period, 1 = valid prior or during, 2 = valid during or after, 3 = valid before, during, and after
	Description string            // description for this account
	LastModTime rlib.JSONDateTime // auto updated
	LastModBy   int64             // user making the mod
	// W2UIChild      w2uiChild `json:"w2ui"`
}

// SearchGLAccountsResponse is the response data to a request for GLAccounts
type SearchGLAccountsResponse struct {
	Status  string      `json:"status"`
	Total   int64       `json:"total"`
	Records []GLAccount `json:"records"`
}

// AcctDetailsForm is the response data to request for a GLAccount
type AcctDetailsForm struct {
	LID         int64
	PLID        int64
	BID         int64
	BUD         rlib.XJSONBud
	RAID        int64
	TCID        int64
	GLNumber    string
	Status      int64
	Name        string
	AcctType    string
	AllowPost   int64
	Description string
	LastModTime rlib.JSONDateTime
	LastModBy   int64
	CreateTS    rlib.JSONDateTime
	CreateBy    int64
}

// AcctSaveForm used save inputs directly
type AcctSaveForm struct {
	LID         int64
	BID         int64
	RAID        int64
	TCID        int64
	GLNumber    string
	Name        string
	AcctType    string
	Description string
	BUD         rlib.XJSONBud
	PLID        int64
	Status      int64
	AllowPost   int64
}

// SaveAcctInput is the input data format for a Save command
type SaveAcctInput struct {
	Cmd      string       `json:"cmd"` // get, save, delete
	Recid    int64        `json:"recid"`
	FormName string       `json:"name"`
	Record   AcctSaveForm `json:"record"`
}

// GetAccountResponse is the response to get details of an account for the requested Account LID
type GetAccountResponse struct {
	Status string          `json:"status"`
	Record AcctDetailsForm `json:"record"`
}

// AcctDeleteForm is struct used to delete Account
type AcctDeleteForm struct {
	LID int64
}

// acctStatus map
var acctStatus = map[int64]string{
	0: "Unknown",
	1: "Inactive",
	2: "Active",
}

// account type
var acctType = map[int64]string{
	0: "Normal Account",
	// 1: "balance for this particular RentalAgreement",
	// 2: "balance for this payor",
	// 3:  "Reserved",
	// 4:  "Reserved",
	// 5:  "Reserved",
	// 6:  "Reserved",
	// 7:  "Reserved",
	// 8:  "Reserved",
	// 9:  "Reserved",
	10: "Default Cash",
	11: "Default General Receivables",
	12: "Default Gross Scheduled Rent",
	13: "Default Loss To Lease",
	14: "Default Vacancy",
	16: "Default Security Deposit",
	17: "Default Owner Equity",
}

// account allow posts
var acctAllowPosts = map[int64]string{
	0: "Summary Account only, do not allow posts to this ledger",
	1: "Allow posts",
}

// getAccountThingJSList sending down list related with accounts info
func getAccountThingJSList() map[string]map[int64]string {
	accountStuff := make(map[string]map[int64]string)

	accountStuff["allowPostList"] = acctAllowPosts
	accountStuff["typeList"] = acctType
	accountStuff["statusList"] = acctStatus

	return accountStuff
}

// SvcAccountsList generates a list of all Accounts with respect of business id specified by d.BID
// wsdoc {
//  @Title Get list of accounts
//  @URL /v1/accountlist/:BUI
//  @Method  GET
//  @Synopsis Get account list
//  @Description Get all General Ledger Account's list for the requested business
//  @Input WebGridSearchRequest
//  @Response AccountListResponse
// wsdoc }
func SvcAccountsList(w http.ResponseWriter, r *http.Request, d *ServiceData) {

	var (
		funcname = "SvcAccountsList"
		g        AccountListResponse
	)
	fmt.Printf("Entered %s\n", funcname)

	// get rentable types for a business
	m := rlib.GetLedgerList(d.BID)
	fmt.Printf("rlib.GetLedgerList returned %d records\n", len(g.Records))

	// append records in ascending order
	var glAccountList []ListedAccount
	for _, acct := range m {
		glAccountList = append(glAccountList,
			ListedAccount{LID: acct.LID, Name: fmt.Sprintf("%s (%s)", acct.GLNumber, acct.Name)},
		)
	}

	// sort based on name, needs version 1.8 later of golang
	sort.Slice(glAccountList, func(i, j int) bool { return glAccountList[i].Name < glAccountList[j].Name })

	g.Records = glAccountList
	g.Total = int64(len(g.Records))
	g.Status = "success"
	SvcWriteResponse(&g, w)
}

// SvcParentAccountsList generates a list of all possible Parent Accounts with respect of business id specified by d.BID
// wsdoc {
//  @Title Get list of parent accounts
//  @URL /v1/parentaccounts/:BUI
//  @Method  GET
//  @Synopsis Get parent account list
//  @Description Get all Parent Account's list for the requested business
//  @Input
//  @Response AccountListResponse
// wsdoc }
func SvcParentAccountsList(w http.ResponseWriter, r *http.Request, d *ServiceData) {

	var (
		funcname = "SvcParentAccountsList"
		g        AccountListResponse
	)
	fmt.Printf("Entered %s\n", funcname)

	// Need to init some internals for Business
	var xbiz rlib.XBusiness
	rlib.InitBizInternals(d.BID, &xbiz)

	// get rentable types for a business
	m := bizlogic.PossibleParentAccounts(d.BID)
	fmt.Printf("bizlogic.PossibleParentAccounts returned %d records\n", len(g.Records))

	// append records in ascending order
	var glAccountList []ListedAccount
	for _, acct := range m {
		glAccountList = append(glAccountList,
			ListedAccount{LID: acct.LID, Name: fmt.Sprintf("%s (%s)", acct.GLNumber, acct.Name)},
		)
	}

	// sort based on name, needs version 1.8 later of golang
	sort.Slice(glAccountList, func(i, j int) bool { return glAccountList[i].Name < glAccountList[j].Name })

	g.Records = glAccountList
	g.Total = int64(len(g.Records))
	g.Status = "success"
	SvcWriteResponse(&g, w)
}

// SvcPostAccountsList generates a list of all Accounts
// that are permissible to use in Assessment/Receipt Rules
// wsdoc {
//  @Title Get list of post accounts
//  @URL /v1/postaccounts/:BUI
//  @Method  GET
//  @Synopsis Get post account list
//  @Description Get all Post Account's list for the requested business
//  @Input
//  @Response AccountListResponse
// wsdoc }
func SvcPostAccountsList(w http.ResponseWriter, r *http.Request, d *ServiceData) {

	var (
		funcname = "SvcPostAccountsList"
		g        AccountListResponse
	)
	fmt.Printf("Entered %s\n", funcname)

	// Need to init some internals for Business
	var xbiz rlib.XBusiness
	rlib.InitBizInternals(d.BID, &xbiz)

	// get rentable types for a business
	m := bizlogic.PossiblePostAccounts(d.BID)
	fmt.Printf("bizlogic.PossiblePostAccounts returned %d records\n", len(g.Records))

	// append records in ascending order
	var glAccountList []ListedAccount
	for _, acct := range m {
		glAccountList = append(glAccountList,
			ListedAccount{LID: acct.LID, Name: fmt.Sprintf("%s (%s)", acct.GLNumber, acct.Name)},
		)
	}

	// sort based on name, needs version 1.8 later of golang
	sort.Slice(glAccountList, func(i, j int) bool { return glAccountList[i].Name < glAccountList[j].Name })

	g.Records = glAccountList
	g.Total = int64(len(g.Records))
	g.Status = "success"
	SvcWriteResponse(&g, w)
}

// SvcSearchHandlerGLAccounts generates a report of all GLAccounts for a the business unit
// called out in d.BID
// wsdoc {
//  @Title  Search General Ledger Accounts
//  @URL /v1/accounts/:BUI
//  @Method POST
//  @Description This service returns a list of General Ledger accounts
//  @Input WebGridSearchRequest
//  @Response SearchGLAccountsResponse
// wsdoc }
func SvcSearchHandlerGLAccounts(w http.ResponseWriter, r *http.Request, d *ServiceData) {

	var (
		funcname = "SvcSearchHandlerGLAccounts"
		p        rlib.GLAccount
		err      error
		g        SearchGLAccountsResponse
	)

	fmt.Printf("Entered %s\n", funcname)

	srch := fmt.Sprintf("BID=%d", d.BID) // default WHERE clause
	order := "GLNumber ASC, Name ASC"    // default ORDER
	q, qw := gridBuildQuery("GLAccount", srch, order, d, &p)

	// set g.Total to the total number of rows of this data...
	g.Total, err = GetRowCount("GLAccount", qw)
	if err != nil {
		SvcGridErrorReturn(w, err, funcname)
		return
	}
	fmt.Printf("db query = %s\n", q)

	rows, err := rlib.RRdb.Dbrr.Query(q)
	defer rows.Close()
	if err != nil {
		SvcGridErrorReturn(w, err, funcname)
		return
	}

	count := 0
	for rows.Next() {
		var p GLAccount
		var q rlib.GLAccount
		rlib.ReadGLAccounts(rows, &q)
		rlib.MigrateStructVals(&q, &p)
		p.Recid = count

		g.Records = append(g.Records, p)

		count++ // update the count only after adding the record
		if count >= d.wsSearchReq.Limit {
			break // if we've added the max number requested, then exit
		}
	}
	err = rows.Err()
	if err != nil {
		SvcGridErrorReturn(w, err, funcname)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	g.Status = "success"
	SvcWriteResponse(&g, w)
}

// ============================================== //
// This following routine is arranging accounts in parent-child
// fashion (according to w2ui subgrid style), as of now just
// disable it, later it will need to be fixed
// ============================================== //

// // SvcSearchHandlerGLAccounts generates a report of all GLAccounts for a the business unit
// // called out in d.BID
// // wsdoc {
// //  @Title  Search General Ledger Accounts
// //	@URL /v1/accounts/:BUI
// //  @Method  GET, POST
// //	@Synopsis Return a list of General Ledger Accounts
// //  @Description This service returns a list of General Ledger accounts
// //	@Input WebGridSearchRequest
// //  @Response SearchGLAccountsResponse
// // wsdoc }
// func SvcSearchHandlerGLAccounts(w http.ResponseWriter, r *http.Request, d *ServiceData) {

// 	var (
// 		funcname = "SvcSearchHandlerGLAccounts"
// 		p        rlib.GLAccount
// 		err      error
// 		g        SearchGLAccountsResponse
// 	)

// 	fmt.Printf("Entered %s\n", funcname)

// 	srch := fmt.Sprintf("BID=%d", d.BID)                 // default WHERE clause
// 	order := "PLID ASC, LID ASC, GLNumber ASC, Name ASC" // default ORDER
// 	q, qw := gridBuildQuery("GLAccount", srch, order, d, &p)

// 	// set g.Total to the total number of rows of this data...
// 	g.Total, err = GetRowCount("GLAccount", qw)
// 	if err != nil {
// 		SvcGridErrorReturn(w, err, funcname)
// 		return
// 	}
// 	fmt.Printf("db query = %s\n", q)

// 	rows, err := rlib.RRdb.Dbrr.Query(q)
// 	defer rows.Close()
// 	if err != nil {
// 		SvcGridErrorReturn(w, err, funcname)
// 		return
// 	}

// 	// this holds LID keys in ascending order
// 	var sortedLIDKeys rlib.Int64Range

// 	// this map holds values LID -> PLID
// 	acctParentMap := make(map[int64]int64)

// 	// account link: LID -> GLAccount
// 	acctMap := make(map[int64]GLAccount)

// 	count := 0
// 	for rows.Next() {
// 		var p GLAccount
// 		var q rlib.GLAccount
// 		rlib.ReadGLAccounts(rows, &q)
// 		rlib.MigrateStructVals(&q, &p)

// 		// map the account with its LID
// 		acctMap[p.LID] = p
// 		// map account's parent account
// 		acctParentMap[p.LID] = p.PLID
// 		// append LID in sorted slice
// 		sortedLIDKeys = append(sortedLIDKeys, p.LID)

// 		count++ // update the count only after adding the record
// 		if count >= d.wsSearchReq.Limit {
// 			break // if we've added the max number requested, then exit
// 		}
// 	}

// 	err = rows.Err()
// 	if err != nil {
// 		SvcGridErrorReturn(w, err, funcname)
// 		return
// 	}

// 	// this holds the list of deleting account from map, after parent-child relation build-up
// 	deleteAcctKeys := []int64{}

// 	// descending order of LID
// 	sort.Sort(sort.Reverse(sortedLIDKeys))

// 	// find child accounts of parent account, fit it in tree
// 	for _, lid := range sortedLIDKeys {

// 		// get parent LID
// 		plid := acctParentMap[lid]

// 		// if this account is at most parent level then keep continue
// 		if plid == 0 {
// 			continue
// 		}

// 		// get parent account
// 		parentAcct, _ := acctMap[plid]

// 		// get account
// 		childAcct := acctMap[lid]

// 		parentAcct.W2UIChild.Children = append(parentAcct.W2UIChild.Children, childAcct)
// 		acctMap[plid] = parentAcct
// 		deleteAcctKeys = append(deleteAcctKeys, lid)
// 	}

// 	// now delete records which has been put as in child of other account
// 	for _, id := range deleteAcctKeys {
// 		delete(acctMap, id)
// 	}

// 	// this holds PLID keys in ascending order
// 	var sortedPLIDKeys rlib.Int64Range

// 	for plid := range acctMap {
// 		sortedPLIDKeys = append(sortedPLIDKeys, plid)
// 	}

// 	// now sort it in ascending order
// 	sort.Sort(sortedPLIDKeys)

// 	// setRecid is internal function to set Recid used in w2ui grid
// 	setRecid := func(acctMap map[int64]GLAccount) {

// 		// recursive routine
// 		// first declare the function signature, so that we can call it recursively
// 		var childAcctRecid func(acct GLAccount, recid int)

// 		childAcctRecid = func(acct GLAccount, recid int) {
// 			if len(acct.W2UIChild.Children) > 0 {
// 				for id, childAcct := range acct.W2UIChild.Children {
// 					recidx := id
// 					// childID would be parentID + incremental id
// 					childID, _ := strconv.Atoi(strconv.Itoa(acct.Recid) + strconv.Itoa(recidx))
// 					childAcct.Recid = childID
// 					acct.W2UIChild.Children[id] = childAcct
// 					childAcctRecid(childAcct, childID)
// 				}
// 				// TODO: what if someone want to see in ascending order
// 			}
// 		}

// 		mostParentCount := 1
// 		for _, plid := range sortedPLIDKeys {
// 			acct := acctMap[plid]
// 			acct.Recid = mostParentCount
// 			acctMap[plid] = acct
// 			childAcctRecid(acct, mostParentCount)
// 			mostParentCount++
// 		}
// 	}

// 	setRecid(acctMap)

// 	// web response
// 	var records []GLAccount
// 	for _, plid := range sortedPLIDKeys {
// 		acct := acctMap[plid]
// 		records = append(records, acct)
// 	}
// 	g.Records = records
// 	g.Total = int64(len(g.Records))

// 	g.Status = "success"
// 	w.Header().Set("Content-Type", "application/json")
// 	SvcWriteResponse(&g, w)
// }

// ======================================================= //

// SvcFormHandlerGLAccounts formats a complete data record for a gl account suitable for use with the w2ui Form
// For this call, we expect the URI to contain the BID and the LID as follows:
//           0    1     2   3
// uri 		/v1/account/BUI/LID
// The server command can be:
//      get
//      save
//      delete
//-----------------------------------------------------------------------------------
func SvcFormHandlerGLAccounts(w http.ResponseWriter, r *http.Request, d *ServiceData) {
	var (
		err      error
		funcname = "SvcFormHandlerGLAccounts"
	)
	fmt.Printf("Entered %s\n", funcname)
	fmt.Printf("Request: %s:  BID = %d,  LID = %d\n", d.wsSearchReq.Cmd, d.BID, d.ID)

	switch d.wsSearchReq.Cmd {
	case "get":
		if d.ID < 0 {
			err = fmt.Errorf("GLAccount ID is required but was not specified")
			SvcGridErrorReturn(w, err, funcname)
			return
		}
		getGLAccount(w, r, d)
		break
	case "save":
		saveGLAccount(w, r, d)
		break
	case "delete":
		deleteGLAccount(w, r, d)
		break
	default:
		err = fmt.Errorf("Unhandled command: %s", d.wsSearchReq.Cmd)
		SvcGridErrorReturn(w, err, funcname)
		return
	}
}

// saveGLAccount returns the requested receipt
// wsdoc {
//  @Title Save GLAccount
//  @URL /v1/account/:BUI/:LID
//  @Method POST
//  @Synopsis Saves a GLAccount details
//  @Description This service saves a GLAccount.  If :LID exists, it will
//  @Description be updated with the information supplied. All fields must
//  @Description be supplied. If LID is 0, then a new GLAccount is created.
//  @Input SaveAcctInput
//  @Response SvcStatusResponse
// wsdoc }
func saveGLAccount(w http.ResponseWriter, r *http.Request, d *ServiceData) {

	var (
		funcname = "saveGLAccount"
		foo      SaveAcctInput
		a        rlib.GLAccount
		err      error
	)

	fmt.Printf("Entered %s\n", funcname)
	fmt.Printf("record data = %s\n", d.data)

	// get data
	data := []byte(d.data)

	if err = json.Unmarshal(data, &foo); err != nil {
		SvcGridErrorReturn(w, err, funcname)
		return
	}

	// migrate foo.Record data to a struct's fields
	rlib.MigrateStructVals(&foo.Record, &a) // the variables that don't need special handling
	fmt.Printf("saveAcct - first migrate: a = %#v\n", a)

	// data validation
	if a.Name == "" {
		err := fmt.Errorf("Provide account name")
		SvcGridErrorReturn(w, err, funcname)
		return
	}
	if a.GLNumber == "" {
		err := fmt.Errorf("Provide value of GLNumber")
		SvcGridErrorReturn(w, err, funcname)
		return
	}

	// save or update
	if a.LID == 0 && d.ID == 0 {

		//-------------------------------------------------------------------
		// check that given name is already exists for business, or GLNumber
		// both name and GLNumber should be unique
		// VALIDATION 2
		//-------------------------------------------------------------------
		existQuery := `SELECT LID FROM GLAccount WHERE {{.WhereClause}};`
		qc := queryClauses{
			"WhereClause": fmt.Sprintf("Name=\"%s\" OR GLNumber=\"%s\"",
				strings.ToLower(a.Name), strings.ToLower(a.GLNumber)),
		}
		q := renderSQLQuery(existQuery, qc)
		fmt.Printf("db query = %s\n", q)
		rows, err := rlib.RRdb.Dbrr.Query(q)
		defer rows.Close()
		if err != nil {
			SvcGridErrorReturn(w, err, funcname)
			return
		}
		for rows.Next() {
			err := fmt.Errorf("GLAccount is already exists for given name or GLNumber")
			SvcGridErrorReturn(w, err, funcname)
			return
		}

		//-------------------------------------------------------------------
		// OK, it's a new account.  Do the bizlogic checks and save...
		//-------------------------------------------------------------------
		errlist := bizlogic.SaveGLAccount(&a)
		if len(errlist) > 0 {
			SvcErrListReturn(w, errlist, funcname)
			return
		}

		//-------------------------------------------------------------------
		// Since it is a new Account, we need a LedgerMarker for it...
		//-------------------------------------------------------------------
		var lm = rlib.LedgerMarker{
			BID:   a.BID,
			LID:   a.LID,
			Dt:    time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC),
			State: rlib.MARKERSTATEORIGIN,
		}
		err = rlib.InsertLedgerMarker(&lm)
		if err != nil {
			e := fmt.Errorf("Error saving Account %s LedgerMarker, Error:= %s", a.Name, err.Error())
			SvcGridErrorReturn(w, e, funcname)
			return
		}
	} else {
		// update existing record
		errlist := bizlogic.SaveGLAccount(&a)
		if len(errlist) > 0 {
			SvcErrListReturn(w, errlist, funcname)
			return
		}
	}

	SvcWriteSuccessResponseWithID(w, a.LID)
}

// which fields needs to be fetched for SQL query for receipts grid
var getAcctQuerySelectFields = selectQueryFields{
	"GLAccount.LID",
	"GLAccount.PLID",
	"GLAccount.RAID",
	"GLAccount.TCID",
	"GLAccount.GLNumber",
	"GLAccount.Status",
	"GLAccount.Name",
	"GLAccount.AcctType",
	"GLAccount.AllowPost",
	"GLAccount.Description",
	"GLAccount.LastModTime",
	"GLAccount.LastModBy",
	"GLAccount.CreateTS",
	"GLAccount.CreateBy",
}

// getGLAccount returns the requested glaccount
// wsdoc {
//  @Title  Get account details
//  @URL /v1/account/:BUI/:LID
//  @Method POST
//  @Synopsis Get details about an account
//  @Description  Return all fields for account :LID
//  @Input WebGridSearchRequest
//  @Response GetAccountResponse
// wsdoc }
func getGLAccount(w http.ResponseWriter, r *http.Request, d *ServiceData) {

	var (
		funcname = "getGLAccount"
		g        GetAccountResponse
		err      error
		order    = `GLAccount.LID ASC`
		whr      = fmt.Sprintf(`GLAccount.BID=%d AND GLAccount.LID=%d`, d.BID, d.ID)
	)

	fmt.Printf("entered %s\n", funcname)

	glQuery := `
	SELECT
		{{.SelectClause}}
	FROM GLAccount
	WHERE {{.WhereClause}}
	ORDER BY {{.OrderClause}};`

	qc := queryClauses{
		"SelectClause": strings.Join(getAcctQuerySelectFields, ","),
		"WhereClause":  whr,
		"OrderClause":  order,
	}

	// get formatted query with substitution of select, where, order clause
	q := renderSQLQuery(glQuery, qc)
	fmt.Printf("db query = %s\n", q)

	// execute the query
	rows, err := rlib.RRdb.Dbrr.Query(q)
	defer rows.Close()
	if err != nil {
		SvcGridErrorReturn(w, err, funcname)
		return
	}

	for rows.Next() {
		var gg AcctDetailsForm
		gg.BID = d.BID
		gg.BUD = getBUDFromBIDList(d.BID)

		err = rows.Scan(&gg.LID, &gg.PLID, &gg.RAID, &gg.TCID, &gg.GLNumber, &gg.Status, &gg.Name, &gg.AcctType, &gg.AllowPost, &gg.Description, &gg.LastModTime, &gg.LastModBy, &gg.CreateTS, &gg.CreateBy)
		if err != nil {
			SvcGridErrorReturn(w, err, funcname)
			return
		}

		g.Record = gg
	}

	// error check
	err = rows.Err()
	if err != nil {
		SvcGridErrorReturn(w, err, funcname)
		return
	}

	g.Status = "success"
	SvcWriteResponse(&g, w)
}

// deleteGLAccount request delete GLAccount from database
// wsdoc {
//  @Title  Delete GLAccount
//  @URL /v1/account/:BUI/:LID
//  @Method  DELETE
//  @Synopsis Delete record for a GL Account
//  @Description  Delete the GL Account for a database and delete its
//  @Description  associated LedgerMarkers.  Use with caution. Only use
//  @Description  this command if you really understand what you're doing.
//  @Input AcctDeleteForm
//  @Response SvcStatusResponse
// wsdoc }
func deleteGLAccount(w http.ResponseWriter, r *http.Request, d *ServiceData) {

	var (
		funcname = "deleteGLAccount"
		del      AcctDeleteForm
	)
	fmt.Printf("Entered %s\n", funcname)
	fmt.Printf("record data = %s\n", d.data)

	if err := json.Unmarshal([]byte(d.data), &del); err != nil {
		SvcGridErrorReturn(w, err, funcname)
		return
	}

	// First check, account exists or not
	gl := rlib.GetLedger(del.LID)
	if gl.LID == 0 {
		err := fmt.Errorf("No such account exists with ID: %d", del.LID)
		SvcGridErrorReturn(w, err, funcname)
		return
	}

	// First, remove LedgerMarkers for this LID
	lm := rlib.GetLatestLedgerMarkerByLID(d.BID, del.LID)
	if lm.State != rlib.MARKERSTATEORIGIN {
		e := fmt.Errorf("This account (LID = %d) cannot be deleted because Ledger Markers exist beyond the origin", del.LID)
		SvcGridErrorReturn(w, e, funcname)
		return
	}
	if err := rlib.DeleteLedgerMarker(lm.LMID); err != nil {
		SvcGridErrorReturn(w, err, funcname)
		return
	}

	if err := rlib.DeleteLedger(del.LID); err != nil {
		SvcGridErrorReturn(w, err, funcname)
		return
	}

	SvcWriteSuccessResponse(w)
}
