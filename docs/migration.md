### Migration

#### Generate

```bash
forgesql migration --generate --name create_posts --environment dev
```

or

```bash
forgesql migration -gn create_posts -e dev
```

#### Up

```bash
❯ forgesql migration --up --environment dev
migrations/dev/1699298774_create_posts.up.sql
```

or

```bash
❯ forgesql migration -ue dev
migrations/dev/1699298774_create_posts.up.sql
```


#### Down

```bash
❯ forgesql migration --down --environment dev
migrations/dev/1699298774_create_posts.down.sql
```

or

```bash
❯ forgesql migration -de dev
migrations/dev/1699298774_create_posts.down.sql
```
