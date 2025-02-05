create table task (
	task_id uuid primary key not null default gen_random_uuid(),
	quantity int not null,
	unit text not null,
	name text not null,
	description text,
	frequency text not null,
	experience_gained int not null,
	is_public boolean not null
);

create table category (
	category_id uuid primary key not null default gen_random_uuid(),
	name text not null
);

create table user_task (
	user_task_id uuid primary key not null default gen_random_uuid(),
	user_id uuid not null,
	task_id uuid not null,
	foreign key (user_id) references "user"(user_id),
	foreign key (task_id) references task(task_id)
);

create table task_completion (
	user_task_id uuid primary key not null,
	complete_timestamp timestamp not null,
	foreign key (user_task_id) references user_task(user_task_id)
);

create table task_category (
	category_id uuid not null,
	task_id uuid not null,
	foreign key (category_id) references category(category_id),
	foreign key (task_id) references task(task_id),
	primary key (category_id, task_id)
);

