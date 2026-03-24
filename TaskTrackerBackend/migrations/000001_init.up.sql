CREATE TABLE "User" (
    id_pk SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL
);


CREATE TABLE Task (
    id_pk SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    owner_id_fk INTEGER NOT NULL,

    CONSTRAINT fk_task_owner FOREIGN KEY (owner_id_fk) REFERENCES "User"(id_pk) ON DELETE CASCADE
);


CREATE TABLE Bug (
    id_pk SERIAL PRIMARY KEY,
    task_id_fk INTEGER NOT NULL,
    severity VARCHAR(50),
    priority VARCHAR(50),
    os VARCHAR(200),
    status VARCHAR(50),
    version_product VARCHAR(50),
    description TEXT,

    created_by_fk INTEGER NOT NULL,
    created_time DATE DEFAULT CURRENT_DATE,

    assigned_to_fk INTEGER,
    assigned_time DATE,

    passed_by_fk INTEGER,
    passed_time DATE,

    accepted_by_fk INTEGER,
    accepted_time DATE,

    playback_description TEXT,
    expected_result TEXT,
    actual_result TEXT,

    CONSTRAINT fk_bug_task FOREIGN KEY (task_id_fk) REFERENCES Task(id_pk) ON DELETE CASCADE,
    CONSTRAINT fk_bug_creator FOREIGN KEY (created_by_fk) REFERENCES "User"(id_pk),
    CONSTRAINT fk_bug_assignee FOREIGN KEY (assigned_to_fk) REFERENCES "User"(id_pk),
    CONSTRAINT fk_bug_tester FOREIGN KEY (passed_by_fk) REFERENCES "User"(id_pk),
    CONSTRAINT fk_bug_approver FOREIGN KEY (accepted_by_fk) REFERENCES "User"(id_pk)
);
