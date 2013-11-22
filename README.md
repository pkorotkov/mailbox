mailbox
=======

Using mailbox library you can easily send full-fledged emails with little effort:

```go
creds := mailbox.NewCredentials("smtp.gmail.com:587")
creds.SetPasswordAuth("my_login", "my_password")

mes := new(mailbox.Message)
mes.From("Pavel", "***@gmail.com").
	To("Dmitry", "***@gmail.com").
	To("Sergey", "***@gmail.com").
	Subject("Meet-up").
	Body("It's high time guys!")
if err := mes.Attach("Agenda.txt"); err != nil {
	log.Fatal(err)
}

if err := SendMessage(creds, mes); err != nil {
	log.Fatal(err)
}
```
