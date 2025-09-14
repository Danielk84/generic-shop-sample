package main

func main() {
	config := NewAppConfig()
	app := NewApp(config)

	app.Run()
}
