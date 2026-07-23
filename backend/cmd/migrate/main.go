package main

func main() {
	app, err := createApp()
	if err != nil {
		panic(err)
	}
	if err := app.MigrationManager.Execute(); err != nil {
		panic(err)
	}
}
