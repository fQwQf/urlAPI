<script setup>
import { ref } from 'vue';
import {Setting} from "@/js/util.js";

const settings = ref()

async function getSetting() {
  settings.value = await Setting("fetchSettings", "azure")
}

async function sendSetting() {
  await Setting("editSettings", "azure", settings.value)
}
</script>

<template>
  <mdui-collapse>
    <mdui-collapse-item rounded>
      <mdui-list-item slot="header" icon="settings_applications" rounded @click="getSetting()">
        Azure OpenAI
        <mdui-icon slot="end-icon" name="keyboard_arrow_down"></mdui-icon>
      </mdui-list-item>
      <mdui-list-item nonclickable>
        <mdui-card variant="outlined">
          <p>这里设置Azure OpenAI的后端API，可用于文字生成，总结等</p>
          <p style="color: #f00;">注意：API地址需要包含完整的deployment路径，格式如：https://your-resource.openai.azure.com/openai/deployments/your-deployment/chat/completions?api-version=2024-02-01</p>
          <mdui-text-field variant="outlined" label="API Key" type="password" toggle-password
                            :value="settings?.api_key || ''"
                            @change="settings.api_key = $event.target.value"></mdui-text-field>
          <mdui-text-field variant="outlined" label="API地址"
                            :value="settings?.endpoint || 'https://{your-resource}.openai.azure.com/openai/deployments/{your-deployment}/chat/completions?api-version=2024-02-01'"
                            @change="settings.endpoint = $event.target.value"></mdui-text-field>
          <p>默认文字生成模型</p>
          <mdui-radio-group :value="settings?.text_model || 'gpt-4o'" style="margin-top: 0"
                            @change="settings.text_model=$event.target.value">
            <mdui-radio value="gpt-4o">GPT-4o</mdui-radio>
            <mdui-radio value="gpt-4o-mini">GPT-4o-mini</mdui-radio>
            <mdui-radio value="gpt-4">GPT-4</mdui-radio>
            <mdui-radio value="gpt-35-turbo">GPT-3.5 Turbo</mdui-radio>
          </mdui-radio-group>
          <p>默认总结模型</p>
          <mdui-radio-group :value="settings?.summary_model || 'gpt-4o-mini'" style="margin-top: 0"
                            @change="settings.summary_model=$event.target.value">
            <mdui-radio value="gpt-4o">GPT-4o</mdui-radio>
            <mdui-radio value="gpt-4o-mini">GPT-4o-mini</mdui-radio>
            <mdui-radio value="gpt-35-turbo">GPT-3.5 Turbo</mdui-radio>
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
