package Commands

func GetResticVersion() {
	runRestic("version")
}
