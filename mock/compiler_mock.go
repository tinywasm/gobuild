package gobuildmock

// FakeCompiler is a mock implementation of the gobuild.Compiler interface.
type FakeCompiler struct {
	CompileErr       error
	CompileCallCount int
	Output           string
	Unobserved       []string
}

// CompileProgram mocks the CompileProgram method.
func (f *FakeCompiler) CompileProgram() error {
	f.CompileCallCount++
	return f.CompileErr
}

// FinalOutputPath mocks the FinalOutputPath method.
func (f *FakeCompiler) FinalOutputPath() string {
	return f.Output
}

// UnobservedFiles mocks the UnobservedFiles method.
func (f *FakeCompiler) UnobservedFiles() []string {
	return f.Unobserved
}
