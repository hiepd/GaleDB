package parser

type (
	Statement interface {
		iStatement()
	}

	Select struct {
		From *From
	}

	From struct {
		TableName string
	}
)

func (*Select) iStatement() {}
func (*From) iStatement()   {}

func NewSelect(from Statement) Statement {
	return &Select{
		From: from.(*From),
	}
}

func NewFrom(tableName string) Statement {
	return &From{
		TableName: tableName,
	}
}
