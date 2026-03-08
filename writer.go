package loading

type writer struct {
	bar *Bar
}

// Write implements [io.Writer].
func (w *writer) Write(p []byte) (int, error) {
	w.bar.mu.Lock()
	defer w.bar.mu.Unlock()

	ansiCursorSave(w.bar.out)
	ansiCursorEnd(w.bar.out, w.bar)
	ansiClearLine(w.bar.out)
	ansiCursorRestore(w.bar.out)

	n, err := w.bar.out.Write(p)
	w.bar.draw(w.bar.out)

	return n, err
}
