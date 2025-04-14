package build
 import(
	"github.com/lariskovski/containy/internal/parser"
	"github.com/lariskovski/containy/internal/instructions"
)

func Build(filepath string){
	fileInstructions, err := parser.ParseFile(filepath)
	if err != nil {
		panic(err)
	}
	// Execute the parsed instructions
	err = instructions.ExecuteInstructions(fileInstructions)
	if err != nil {
		panic(err)
	}

}