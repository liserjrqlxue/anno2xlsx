package main

//go:generate sed "s|118b09d39a5d3ecd56f9bd4f351dd6d6|ppCode|g" -i global.go
//go:generate sed "s|afe0ba6b37644a81cedc2364ee414c7f|nbCode|g" -i global.go
//go:generate sed "s|6d276bc509883dbafe05be835ad243d7|hlCode|g" -i global.go
//go:generate sed "s|b7ea138a9842cbb832271bdcf4478310|sfCode|g" -i global.go
//go:generate vb -ldflags "-w -s"
//go:generate git checkout global.go
