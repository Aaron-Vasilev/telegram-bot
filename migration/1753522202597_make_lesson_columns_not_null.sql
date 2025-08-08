ALTER TABLE yoga.lesson
  alter column description set not null,
  alter column description set default '',
  alter column max set not null;
