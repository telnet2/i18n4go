package rewrite_package_test

import (
	. "github.com/maximilien/i18n4cf/integration/test_helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var _ = Describe("rewrite-package -d dirname -r", func() {
	var (
		outputDir         string
		rootPath          string
		fixturesPath      string
		inputFilesPath    string
		expectedFilesPath string
	)

	BeforeEach(func() {
		dir, err := os.Getwd()
		Ω(err).ShouldNot(HaveOccurred())
		rootPath = filepath.Join(dir, "..", "..")
		outputDir = filepath.Join(rootPath, "tmp")

		fixturesPath = filepath.Join("..", "..", "test_fixtures", "rewrite_package")
		inputFilesPath = filepath.Join(fixturesPath, "f_option", "input_files")
		expectedFilesPath = filepath.Join(fixturesPath, "f_option", "expected_output")

		session := Runi18n(
			"-rewrite-package",
			"-d", inputFilesPath,
			"-o", outputDir,
			"-r",
			"-v",
		)

		Ω(session.ExitCode()).Should(Equal(0))
	})

	It("adds T() callExprs wrapping string literals", func() {
		expectedOutputFile := filepath.Join(expectedFilesPath, "test.go")
		bytes, err := ioutil.ReadFile(expectedOutputFile)
		Ω(err).ShouldNot(HaveOccurred())

		expectedOutput := string(bytes)

		generatedOutputFile := filepath.Join(outputDir, "test.go")
		bytes, err = ioutil.ReadFile(generatedOutputFile)
		Ω(err).ShouldNot(HaveOccurred())

		actualOutput := string(bytes)
		Ω(actualOutput).Should(Equal(expectedOutput))
	})

	It("recurses to files in nested dirs", func() {
		expectedOutputFile := filepath.Join(expectedFilesPath, "nested_dir", "test.go")
		bytes, err := ioutil.ReadFile(expectedOutputFile)
		Ω(err).ShouldNot(HaveOccurred())

		expectedOutput := string(bytes)

		generatedOutputFile := filepath.Join(outputDir, "nested_dir", "test.go")
		bytes, err = ioutil.ReadFile(generatedOutputFile)
		Ω(err).ShouldNot(HaveOccurred())

		actualOutput := string(bytes)
		Ω(actualOutput).Should(Equal(expectedOutput))
	})

	It("adds a i18n_init.go per package", func() {
		initFile := filepath.Join(outputDir, "i18n_init.go")
		expectedBytes, err := ioutil.ReadFile(initFile)
		Ω(err).ShouldNot(HaveOccurred())
		expected := strings.TrimSpace(string(expectedBytes))

		expectedInitFile := filepath.Join(expectedFilesPath, "i18n_init.go")
		actualBytes, err := ioutil.ReadFile(expectedInitFile)
		Ω(err).ShouldNot(HaveOccurred())
		actual := strings.TrimSpace(string(actualBytes))

		Ω(actual).Should(Equal(expected))

		initFile = filepath.Join(outputDir, "nested_dir", "i18n_init.go")
		expectedBytes, err = ioutil.ReadFile(initFile)
		Ω(err).ShouldNot(HaveOccurred())
		expected = strings.TrimSpace(string(expectedBytes))

		expectedInitFile = filepath.Join(expectedFilesPath, "nested_dir", "i18n_init.go")
		actualBytes, err = ioutil.ReadFile(expectedInitFile)
		Ω(err).ShouldNot(HaveOccurred())
		actual = strings.TrimSpace(string(actualBytes))

		Ω(actual).Should(Equal(expected))
	})

	It("does not translate test files", func() {
		_, doesFileExistErr := os.Stat(filepath.Join(outputDir, "a_really_bad_test.go"))
		Ω(os.IsNotExist(doesFileExistErr)).Should(BeTrue())
	})
})