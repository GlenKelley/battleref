package tournament

// SchemaMigrations defines a sequence of idompotent changes to the database schema, starting from an clean sqlite database
var SchemaMigrations = map[string][]string {
	"0.0.1":[]string{"create table if not exists schema_log(version text not null primary key, date_applied timestamp not null default current_timestamp);"},
	"0.0.2":[]string{
		"create table if not exists user (name text not null primary key, public_key text not null, date_created timestamp not null default current_timestamp)",
		"create table if not exists submission (commithash text not null primary key, name text not null, category text not null, date_created timestamp not null default current_timestamp)",
		"create table if not exists map (name text primary key, source text not null)",
		"create table if not exists match (category text not null, p1 text not null, p2 text not null, map text not null, result text not null, unique (p1, p2, map))",
	},
}

