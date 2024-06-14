package transactions

type ofxTransaction struct {
	Date        string `xml:"DTPOSTED"`
	Description string `xml:"NAME"`
	Amount      string `xml:"TRNAMT"`
}

type ofx struct {
	Transactions []ofxTransaction `xml:"BANKMSGSRSV1>STMTTRNRS>STMTRS>BANKTRANLIST>STMTTRN"`
}
