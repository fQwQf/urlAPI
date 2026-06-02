<script setup>

import OpenAI from "@/Components/Backend/OpenAI.vue";
import Alibaba from "@/Components/Backend/Alibaba.vue";
import Other from "@/Components/Backend/Other.vue";
import Deepseek from "@/Components/Backend/Deepseek.vue";
import Anthropic from "@/Components/Backend/Anthropic.vue";
import Gemini from "@/Components/Backend/Gemini.vue";
import Azure from "@/Components/Backend/Azure.vue";
import Moonshot from "@/Components/Backend/Moonshot.vue";
import Context from "@/Components/Backend/Context.vue";
import WebAPI from "@/Components/Backend/WebAPI.vue";
import {onMounted, inject} from "vue";

const title = inject("title");

onMounted(()=>{
  title.value = "接口设置"
})
</script>

<template>
  <mdui-layout-main>
    <OpenAI></OpenAI>
    <Alibaba></Alibaba>
    <Deepseek></Deepseek>
    <Anthropic></Anthropic>
    <Gemini></Gemini>
    <Azure></Azure>
    <Moonshot></Moonshot>
    <Other></Other>
    <Context></Context>
    <WebAPI></WebAPI>
  </mdui-layout-main>
</template>

<style scoped>

</style>
