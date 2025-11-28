/// <reference types="vite/client" />

interface ImportMetaEnv {
	readonly BACKEND_BASE_URL: string
}

interface ImportMeta {
	readonly env: ImportMetaEnv
}
