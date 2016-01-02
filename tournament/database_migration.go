package tournament

// SchemaMigrations defines a sequence of idompotent changes to the database schema, starting from an clean sqlite database
var SchemaMigrations = map[string][]string{
	"0.0.1": []string{"create table if not exists schema_log(version text not null primary key, date_applied timestamp not null default current_timestamp)"},
	"0.0.2": []string{
		"create table if not exists user (name text not null primary key, public_key integer not null, date_created timestamp not null default current_timestamp)",
		"create table if not exists pkey (id integer primary key autoincrement, key text not null, unique(key))",
		"create table if not exists submission (commithash text not null, name text not null, category text not null, date_created timestamp not null default current_timestamp, unique (commithash, name))",
		"create table if not exists map (name text, source text not null, category text not null, unique(name, category))",
		"create table if not exists match (id integer primary key, category text not null, player1 text not null, player2 text not null, commit1 text not null, commit2 text not null, map text not null, result text not null, created timestamp not null default current_timestamp, updated timestamp default null, replay blob default null, unique (category, map, player1, player2, commit1, commit2))",
	},
}
