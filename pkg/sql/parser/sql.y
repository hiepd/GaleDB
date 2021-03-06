%{
package parser

func setParseTree(yylex yyLexer, stmt Statement) {
  yylex.(*Lexer).ParseTree = stmt
}
%}
    /*symbolic tokens*/
%union {
    str string
    num int
    statement Statement
    strs []string
    cond *Condition
    conds []*Condition
    where *Where
}

%token LEX_ERROR
%token <str> NAME
%token <num> NUMBER
%token STRING
%token INTNUM APPROXNUM

    /* operators */
%left OR
%left AND
%left NOT
%left <str> RELATION
%left OPERATOR
%nonassoc '.'

    /*literal keyword tokens*/
%token <str> ASTERISK ALL AMMSC ANY ASC AS AUTHORIZATION AVG BETWEEN BY
%token <str> CHARACTER CHECK CLOSE COMMIT CONTINUE CREATE CURRENT COMMA
%token <str> CURSOR DECIMAL DECLARE DEFAULT DELETE DESC DISTINCT DOUBLE
%token <str> ESCAPE EXISTS FETCH FLOAT FOR FOREIGN FOUND FROM GOTO
%token <str> GRANT GROUP HAVING IN INDICATOR INSERT INTEGER INTO IS MIN MAX
%token <str> KEY LANGUAGE LIKE NULLX NUMERIC OF ON OPEN OPTION
%token <str> ORDER PARAMETER PRECISION PRIMARY PRIVILEGES PROCEDURE
%token <str> PUBLIC REAL REFERENCES ROLLBACK SCHEMA SELECT SET
%token <str> SMALLINT SOME SQLCODE SQLERROR SUM TABLE TO UNION
%token <str> UNIQUE UPDATE USER VALUES VIEW WHENEVER WHERE WITH WORK

%type <str> table column
%type <strs> column_commalist
%type <cond> condition
%type <conds> condition_list
%type <where> where_clause

%type <statement> sql
%type <statement> manipulative_statement select_statement from_clause

%start sql

%%

sql: 
    manipulative_statement { setParseTree(yylex, $1) }
    ;

    /* schema */

schema: 
        CREATE SCHEMA AUTHORIZATION user
    opt_schema_element_list
    ;

opt_schema_element_list: 
        /* empty */
    | schema_element_list
    ;

schema_element_list:
        schema_element
    | schema_element_list schema_element
    ;

schema_element:
        base_table_def
    | view_def
    ;

base_table_def:
        CREATE TABLE table '(' base_table_element_commalist ')'
    ;

base_table_element_commalist:
        base_table_element
    | base_table_element_commalist ',' base_table_element
    ;

base_table_element: 
        column_def
    | table_constraint_def
    ;

column_def:
        column data_type column_def_opt_list
    ;

column_def_opt_list:
        /* empty */
    | column_def_opt_list column_def_opt
    ;

column_def_opt:
        NOT NULLX
    | NOT NULLX UNIQUE
    ;

table_constraint_def:
        UNIQUE '(' column_commalist ')'
    ;

column_commalist:
        column { $$ = []string{$1} }
    | column_commalist COMMA column { $$ = append($1, $3) }
    ;

view_def:
        CREATE VIEW table opt_column_commalist
    ;

column:
        NAME
     {
     	$$ = $1
     }
    ;

opt_column_commalist:
        /* empty */
    | '(' column_commalist ')'
    ;

    /* manipulative statement */

manipulative_statement:
    select_statement { $$ = $1 }
    ;

close_statement:
        CLOSE
    ;

commit_statement:
        COMMIT WORK
    ;

insert_statement:
        INSERT INTO table opt_column_commalist values_or_query_spec
    ;

values_or_query_spec:
        VALUES '(' insert_atom_commalist ')'
    ;

insert_atom_commalist:
        insert_atom
    | insert_atom_commalist ',' insert_atom
    ;

insert_atom:
        atom
    | NULLX
    ;

open_statement:
        OPEN
    ;

rollback_statement:
        ROLLBACK
    ;

select_statement:
    	/*  1       2       	3   				4		*/
        SELECT column_commalist from_clause where_clause
        { 
            $$ = NewSelect($2, $3, $4)
        }
    |	SELECT column_commalist from_clause
        {
            $$ = NewSelect($2, $3, nil)
        }
    ;

from_clause:
    /*    1    2  */
        FROM table
        {
            $$ = NewFrom($2)
        }
    ;

where_clause:
		WHERE condition_list
		{
			$$ = NewWhere($2)
		}
	;

condition_list:
	condition { $$ = []*Condition{$1} }
	| condition_list AND condition { $$ = append($1, $3) }
	;

condition:
	NAME RELATION NAME { $$ = NewCondition($2, $1, $3) }
	| NAME RELATION NUMBER { $$ = NewCondition($2, $1, $3) }
	| NUMBER RELATION NUMBER { $$ = NewCondition($2, $1, $3) }
	;

atom:
        parameter_ref
    | literal
    | USER
    ;

parameter_ref:
        parameter
    | parameter parameter
    | parameter INDICATOR parameter
    ;

literal:
        STRING
    | INTNUM
    | APPROXNUM
    ;

table: 
        NAME { $$ = $1 }
    | NAME '.' NAME { $$ = $1 }
    ;

data_type:
        CHARACTER
    | NUMERIC
    | FLOAT
    ;

parameter: 
        PARAMETER
    ;

user: 
        NAME
    ;

%%