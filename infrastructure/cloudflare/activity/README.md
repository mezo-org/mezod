# Activity

This module contains the implementation of the Mezo chain activity component.
This component is written and deployed as a Cloudflare Worker.

### Prerequisites

Initially, you need to install the dependencies by running
```shell
npm install
```

### Development

First, apply local database migrations by running:
```shell
npm run migrations:local:apply
```

You can verify applied migrations by running:
```shell
npm run migrations:local:list
```

Once done, you can start the development server by running:
```shell
npm run dev
```

The component will be available at `http://localhost:8001`. Code changes
are hot-reloaded by Wrangler.

### Deployment

First, apply database migrations on the staging environment by running:
```shell
npm run migrations:staging:apply
```

You can verify applied migrations by running:
```shell
npm run migrations:staging:list
```

Once done, you can deploy the component to the staging environment by running:
```shell
npm run deploy:staging
```
