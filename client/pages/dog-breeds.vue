<template>
  <div class="flex flex-col gap-4 items-center">
    <div class="max-w-lg">
      Upload an image, and the classifier will tell you what breed of dog is in the image and the probability it thinks!
    </div>
    <input type="file" class="file-input file-input-bordered w-full max-w-xs" @change="postImage">
    <div v-if="breed" class="text-center flex flex-col items-center">
      <img v-if="imagePath" :src="imagePath" alt="Uploaded Image" class="w-48">
      <div class="text-2xl">
        {{ breed }}
      </div>
      <div class="text-xl">
        {{ probability }}
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
const breed = ref<string | null>(null)
const probability = ref<string | null>(null)
const imagePath = ref<string | null>(null)

async function postImage (event: Event) {
  const target = event.target as HTMLInputElement
  const image = target.files?.item(0)

  if (!image) { return }

  if (image.size >= 5e6) {
    alert('Image is larger than 5 MB!')
    return
  }

  const response = await fetch('/api/models/predictions/resnet50_dog_breeds', {
    method: 'POST',
    body: image
  })

  const result = await response.json()
  imagePath.value = URL.createObjectURL(image)
  breed.value = Object.keys(result)[0]
  probability.value = result[breed.value]
}
</script>
