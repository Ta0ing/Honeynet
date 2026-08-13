Dim FirstDigest
Dim Digest 
Digest= "01234567890123456"

function Validate(RandomData)
	Digest = "01234567890123456"
	On Error Resume Next
	Dim TheForm
	Set TheForm = Document.forms("form1")
	Set snID = Document.form1.snID
	Set DigestID = Document.form1.DigestID
	ePass.GetLibVersion
	ePass.OpenDevice 1, ""
	dim results
	results = "01234567890123456"
	results = ePass.GetStrProperty(7, 0, 0)
	ePass.VerifyPIN 0, CStr(TheForm.UserPIN.Value)
	ePass.ChangeDir &H300, 0, "ASP_DEMO"
	ePass.OpenFile 0, 1
	Digest = ePass.HashToken (1, 2,RandomData)
	DigestID.Value = Digest 
	snID.Value = results
	ePass.CloseDevice
End function
