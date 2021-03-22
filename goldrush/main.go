package main

func main() {
	client := NewClient()
	licenser := NewLicenser(client)
	treasurer := NewTreasurer(client)
	digger := NewDigger(client, licenser, treasurer)
	explorer := NewExplorer(client, digger)

	explorer.Run()
}
