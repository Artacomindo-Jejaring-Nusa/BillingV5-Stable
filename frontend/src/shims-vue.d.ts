/* eslint-disable @typescript-eslint/no-explicit-any, @typescript-eslint/no-empty-object-type */
declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

declare module '@/*' {
  const value: any
  export default value
}

declare module '*.mp3' {
  const src: string
  export default src
}

declare module 'vue3-emoji-picker' {
  import type { DefineComponent } from 'vue'
  const EmojiPicker: DefineComponent<any, any, any>
  export default EmojiPicker
}