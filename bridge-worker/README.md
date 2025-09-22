Bridge-worker
=============

# Supabase migrations

## Install the supabase CLI

Migrations are managed via the supabase CLI, instructions on how to install it can be found here: https://supabase.com/docs/guides/local-development?queryGroups=package-manager&package-manager=npm

## Run the migration

You will need the project reference and an access token. Both can be found from the supabase UI. Then run
the following make rule:

```
SUPABASE_ACCESS_TOKEN=<sbp_NNNNNNN> SUPABASE_PROJECT_REF=<PROJECT_REF> make supabase-deploy
```
