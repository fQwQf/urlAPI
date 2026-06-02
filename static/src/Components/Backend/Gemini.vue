<script setup>
import { ref } from 'vue';
import {Setting} from "@/js/util.js";

const settings = ref()

async function getSetting() {
  settings.value = await Setting("fetchSettings", "gemini")
}

async function sendSetting() {
  await Setting("editSettings", "gemini", settings.value)
}
</script>

<template>
  <mdui-collapse>
    <mdui-collapse-item rounded>
      <mdui-list-item slot="header" icon="settings_applications" rounded @click="getSetting()">
        Google Gemini
        <mdui-icon slot="end-icon" name="keyboard_arrow_down"></mdui-icon>
      </mdui-list-item>
      <mdui-list-item nonclickable>
        <mdui-card variant="outlined">
          <p>这里设置Google Gemini的后端API，可用于文字生成，总结等</p>
          <mdui-text-field variant="outlined" label="API Key" type="password" toggle-password
                            :value="settings?.api_key || ''"
                            @change="settings.api_key = $event.target.value"></mdui-text-field>
          <mdui-text-field variant="outlined" label="API地址"
                            :value="settings?.endpoint || 'https://generativelanguage.googleapis.com/v1beta/models'"
                            @change="settings.endpoint = $event.target.value"></mdui-text-field>
          <p>默认文字生成模型</p>
          <mdui-radio-group :value="settings?.text_model || 'gemini-2.0-flash'" style="margin-top: 0"
                            @change="settings.text_model=$event.target.value">
            <mdui-radio value="gemini-2.0-flash">Gemini 2.0 Flash</mdui-radio>
            <mdui-radio value="gemini-2.0-flash-lite">Gemini 2.0 Flash Lite</mdui-radio>
            <mdui-radio value="gemini-2.0-pro-exp-02-05">Gemini 2.0 Pro Exp</mdui-radio>
            <mdui-radio value="gemini-1.5-pro">Gemini 1.5 Pro</mdui-radio>
            <mdui-radio value="gemini-1.5-flash">Gemini 1.5 Flash</mdui-radio>
          </mdui-radio-group>
          <p>默认总结模型</p>
          <mdui-radio-group :value="settings?.summary_model || 'gemini-2.0-flash-lite'" style="margin-top: 0"
                            @change="settings.summary_model=$event.target.value">
            <mdui-radio value="gemini-2.0-flash">Gemini 2.0 Flash</mdui-radio>
            <mdui-radio value="gemini-2.0-flash-lite">Gemini 2.0 Flash Lite</mdui-radio>
            <mdui-radio value="gemini-1.5-flash">Gemini 1.5 Flash</mdui-radio>
          </mdui-radio-group>
          <p>温度参数 (0-1)</p>
          <mdui-slider :value="settings?.temperature || 1" min="0" max="1" step="0.1"
                        @change="settings.temperature = $event.target.value"></mdui-slider>
          <p>启用状态</p>
          <mdui-switch :checked="settings?.enabled !== false"
                        @change="settings.enabled = $event.target.checked"></mdui-switch>
          <mdui-button @click="sendSetting()">确认</mdui-button>
        </mdui-card>
      </mdui-list-item>
    </mdui-collapse-item>
  </mdui-collapse>
</template>

<style scoped>

</style>
