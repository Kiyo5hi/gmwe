<template>
  <div class="h-full bg-green-500 dark:bg-gray-900 shadow-lg rounded p-3 flex flex-row items-center">
    <div class="h-full group relative flex">
      <img class="max-h-full w-auto my-auto block rounded" :src="props.cover">
      <div
        class="absolute bg-black rounded bg-opacity-0 group-hover:bg-opacity-60 w-full h-full top-0 flex items-center group-hover:opacity-100 transition justify-evenly"
      >
        <button
          class="hover:scale-110 text-white opacity-0 transform translate-y-3 group-hover:translate-y-0 group-hover:opacity-100 transition"
          @click="play"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="40"
            height="40"
            fill="currentColor"
            class="bi bi-play-circle-fill"
            viewBox="0 0 16 16"
          >
            <path
              :class="{ hidden: isPlaying }"
              d="M16 8A8 8 0 1 1 0 8a8 8 0 0 1 16 0zM6.79 5.093A.5.5 0 0 0 6 5.5v5a.5.5 0 0 0 .79.407l3.5-2.5a.5.5 0 0 0 0-.814l-3.5-2.5z"
            />
            <path
              :class="{ hidden: !isPlaying }"
              d="M16 8A8 8 0 1 1 0 8a8 8 0 0 1 16 0zM6.25 5C5.56 5 5 5.56 5 6.25v3.5a1.25 1.25 0 1 0 2.5 0v-3.5C7.5 5.56 6.94 5 6.25 5zm3.5 0c-.69 0-1.25.56-1.25 1.25v3.5a1.25 1.25 0 1 0 2.5 0v-3.5C11 5.56 10.44 5 9.75 5z"
            />
          </svg>
        </button>
      </div>
    </div>
    <div class="p-5 m-auto">
      <h3 class="text-white text-lg">
        {{ props.name }}
      </h3>
      <p class="text-gray-800 dark:text-gray-400">
        {{ props.artist }}
      </p>
    </div>
  </div>
</template>

<script lang="ts" setup>
const props = defineProps({
  name: {
    type: String,
    required: true
  },
  artist: {
    type: String,
    required: true
  },
  url: {
    type: String,
    required: true
  },
  cover: {
    type: String,
    required: true
  }
})

const music = new Audio(props.url)
const isPlaying = ref(false)

function play () {
  if (!isPlaying.value) {
    music.play()
  } else {
    music.pause()
  }
  isPlaying.value = !isPlaying.value
}
</script>
