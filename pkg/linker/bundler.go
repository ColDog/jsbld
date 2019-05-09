package linker


func StandardBundler(b *Bundle) error {
	chunks := map[string]*Chunk{}
	for _, entrypoint := range b.Entrypoints {
		chunks[entrypoint] = &Chunk{
			Entrypoint: entrypoint,
			Files: Files{},
		}
	}
	for name, file := range b.Files {
		for _, entrypoint := range file.Entrypoints {
			chunks[entrypoint].Files[name] = file
		}
	}
	for _, c := range chunks {
		b.Chunks = append(b.Chunks, c)
	}
	return nil
}

