"use strict";
/*global
    GridMoneyFormat, number_format
*/

function buildStatementsElements() {

    //------------------------------------------------------------------------
    //          stmtGrid  -  THE LIST OF ALL RENTAL AGREEMENTS
    //------------------------------------------------------------------------
    $().w2grid({
        name: 'stmtGrid',
        url: '/v1/stmt',
        multiSelect: false,
        postData: {searchDtStart: app.D1, searchDtStop: app.D2},
        show: {
            toolbar         : true,
            footer          : true,
            toolbarAdd      : false,   // indicates if toolbar add new button is visible
            toolbarDelete   : false,   // indicates if toolbar delete button is visible
            toolbarSave     : false,   // indicates if toolbar save button is visible
            selectColumn    : false,
            expandColumn    : false,
            toolbarEdit     : false,
            toolbarSearch   : false,
            toolbarInput    : true,
            searchAll       : false,
            toolbarReload   : true,
            toolbarColumns  : true,
        },
        columns: [
            {field: 'recid', hidden: true,  caption: 'recid',            size: '40px',  sortable: true},
            {field: 'BID',   hidden: true,  caption: 'BID',              size: '40px',  sortable: true},
            {field: 'RAID',  hidden: false, caption: 'Rental Agreement', size: '200px', sortable: true},
            {field: 'Payors',hidden: false, caption: 'Payors',           size: '250px', sortable: true},
        ],
        onClick: function(event) {
            event.onComplete = function () {
                var yes_args = [this, event.recid],
                    no_args = [this],
                    no_callBack = function(grid) {
                        grid.select(app.last.grid_sel_recid);
                        return false;
                    },
                    yes_callBack = function(grid, recid) {
                        app.last.grid_sel_recid = parseInt(recid);

                        // keep highlighting current row in any case
                        grid.select(app.last.grid_sel_recid);

                        var rec = grid.get(recid);
                        console.log( 'BID = ' + rec.BID + ',   RAID = ' + rec.RAID);
                        var d = new Date();  // we'll use today for time-sensitive data
                        setToStmtForm(rec.BID, rec.RAID, d,d);
                    };

                // warn user if form content has been changed
                form_dirty_alert(yes_callBack, no_callBack, yes_args, no_args);
            };
        },
    });

    addDateNavToToolbar('stmt');

    //------------------------------------------------------------------------
    //  stmtDetailForm
    //------------------------------------------------------------------------
    $().w2form({
        name: 'stmtDetailForm',
        style: 'border: 0px; background-color: transparent;',
        header: 'Statement Detail',
        url: '/v1/stmtinfo',
        formURL: '/html/formstmtdet.html',
        toolbar: {
            items: [
                { id: 'btnNotes', type: 'button', icon: 'fa fa-sticky-note-o' },
                { id: 'bt3', type: 'spacer' },
                { id: 'btnClose', type: 'button', icon: 'fa fa-times' },
            ],
            onClick: function (event) {
                switch(event.target) {
                case 'btnClose':
                    var no_callBack = function() { return false; },
                        yes_callBack = function() {
                            w2ui.toplayout.hide('right',true);
                            w2ui.stmtGrid.render();
                        };
                    form_dirty_alert(yes_callBack, no_callBack);
                    break;
                }
            },
        },
        fields: [
            { field: 'recid', type: 'int', required: false, html: {page: 0, column: 0 } },
            { field: 'RAID', type: 'int', required: false, html: {  page: 0, column: 0 } },
            { field: 'BID', type: 'int', required: false, html: { page: 0, column: 0 } },
            { field: 'Balance', type: 'float', required: false, html: { page: 0, column: 0 }, render: 'money' },
            { field: 'Payors', type: 'text', required: false, html: { page: 0, column: 0 } },
            { field: 'AgreementStart', type: 'date', required: false, html: { page: 0, column: 0 } },
            { field: 'AgreementStop', type: 'date', required: false, html: { page: 0, column: 0 } },
            { field: 'PossessionStart', type: 'date', required: false, html: { page: 0, column: 0 } },
            { field: 'PossessionStop', type: 'date', required: false, html: { page: 0, column: 0 } },
            { field: 'RentStart', type: 'date', required: false, html: { page: 0, column: 0 } },
            { field: 'RentStop', type: 'date', required: false, html: { page: 0, column: 0 } },
        ],
        onRefresh: function(event) {
            event.onComplete = function() {
                var x = document.getElementById("bannerRAID");
                if (x !== null) {
                    x.innerHTML = '' + this.record.RAID;
                }
                x = document.getElementById("bannerPayors");
                if (x !== null) {
                    x.innerHTML = '' + this.record.Payors;
                }
                x = document.getElementById("RentalAgreementDates");
                if (x !== null) {
                    x.innerHTML = '' + this.record.AgreementStart + ' - ' + this.record.AgreementStop;
                }
                x = document.getElementById("PossessionDates");
                if (x !== null) {
                    x.innerHTML = '' + this.record.PossessionStart + ' - ' + this.record.PossessionStop;
                }
                x = document.getElementById("RentDates");
                if (x !== null) {
                    x.innerHTML = '' + this.record.RentStart + ' - ' + this.record.RentStop;
                }
                x = document.getElementById("CurrentStatementBalance");
                if (x !== null) {
                    x.innerHTML = '$ ' + number_format(this.record.Balance ,2);
                }
            };
        },

    });

    //------------------------------------------------------------------------
    //  stmtDetailGrid  -  lists all the assessments and receipts for
    //                     the selected Rental Agreement from the stmtGrid
    //------------------------------------------------------------------------
    $().w2grid({
        name: 'stmtDetailGrid',
        url: '/v1/stmtdetail',
        multiSelect: false,
        postData: {searchDtStart: app.D1, searchDtStop: app.D2},
        show: {
            toolbar         : true,
            footer          : true,
            toolbarAdd      : false,   // indicates if toolbar add new button is visible
            toolbarDelete   : false,   // indicates if toolbar delete button is visible
            toolbarSave     : false,   // indicates if toolbar save button is visible
            selectColumn    : false,
            expandColumn    : false,
            toolbarEdit     : false,
            toolbarSearch   : false,
            toolbarInput    : false,   // the text area for searches
            searchAll       : false,
            toolbarReload   : false,
            toolbarColumns  : false,
        },
        columns: [
            {field: 'recid',        caption: 'recid',       size: '35px',  sortable: true, hidden: true},
            {field: 'Dt',           caption: 'Date',        size: '75px',  sortable: true},
            {field: 'ID',           caption: 'ID',          size: '80px', sortable: true},
            {field: 'RentableName', caption: app.sRentable, size: '30%', sortable: true},
            {field: 'Descr',        caption: 'Description', size: '60%', sortable: true},
            {field: 'AsmtAmount',   caption: 'Assessment',  size: '90px', sortable: true, style: 'text-align: right',
                    render: function (record,index,col_index) { return stmtRenderHandler(record,index,col_index,record.AsmtAmount); },
            },
            {field: 'RcptAmount',   caption: 'Receipt',     size: '90px', sortable: true, style: 'text-align: right',
                    render: function (record,index,col_index) { return stmtRenderHandler(record,index,col_index,record.RcptAmount); },
            },
            {field: 'Balance',      caption: 'Balance',     size: '90px', sortable: true, style: 'text-align: right',
                    render: function (record,index,col_index) { return stmtRenderHandler(record,index,col_index,record.Balance); },
            },
        ],
    });

    addDateNavToToolbar('stmtDetail');

    //------------------------------------------------------------------------
    //  stmtlayout - The layout to contain the stmtForm and stmtDetailGrid
    //               top  - stmtForm
    //               main - stmtDetailGrid
    //------------------------------------------------------------------------
    $().w2layout({
        name: 'stmtLayout',
        padding: 0,
        panels: [
            { type: 'left',    size: '30%', hidden: true },
            { type: 'top',     size: 250,   hidden: false, content: 'top',  resizable: true, style: app.pstyle },
            { type: 'main',    size: '70%', hidden: false, content: 'main', resizable: true, style: app.pstyle },
            { type: 'preview', size: 0,     hidden: true,  content: 'PREVIEW'  },
            { type: 'bottom',  size: 0,     hidden: true },
            { type: 'right',   size: 0,     hidden: true }
        ]
    });
}


function stmtRenderHandler(record,index,col_index,amt) {
    if (record.Descr.includes("Closing Balance") && Math.abs(amt) < 0.001) {
        return '$ 0.00';
    }
    return GridMoneyFormat(amt); 
}

//-----------------------------------------------------------------------------
// setToStmtForm -  enable the Statement form in toplayout.  Also, set
//                the forms url and request data from the server
// @params
//   bid = business id (or the BUD)
//  raid = Rental Agreement ID
// d1,d2 = date range to use
//-----------------------------------------------------------------------------
function setToStmtForm(bid, raid, d1,d2) {
    if (raid > 0) {
        w2ui.stmtDetailGrid.url = '/v1/stmtdetail/' + bid + '/' + raid;
        w2ui.stmtDetailForm.url = '/v1/stmtinfo/' + bid + '/' + raid;
        w2ui.stmtDetailForm.request();

        w2ui.toplayout.content('right', w2ui.stmtLayout);
        w2ui.toplayout.show('right', true);
        w2ui.toplayout.sizeTo('right', 770);
        w2ui.toplayout.render();
        app.new_form_rec = false; // mark as record exists
        app.form_is_dirty = false; // mark as no changes yet
    }
}

//-----------------------------------------------------------------------------
// createStmtForm - add the grid and form to the statement layout.  I'm not
//      sure why this is necessary. But if I put this grid and form directly
//      into the layout when it gets created, they do not work correctly.
// @params
//-----------------------------------------------------------------------------
function createStmtForm() {
    w2ui.stmtLayout.content('top',w2ui.stmtDetailForm);
    w2ui.stmtLayout.content('main',w2ui.stmtDetailGrid);
}
