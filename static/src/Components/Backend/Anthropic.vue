<script setup>
import { ref } from 'vue';
import {Setting} from "@/js/util.js";

const settings = ref()

async function getSetting() {
  settings.value = await Setting("fetchSettings", "anthropic")
}

async function sendSetting() {
  await Setting("editSettings", "anthropic", settings.value)
}
</script>

<template>
  <mdui-collapse>
    <mdui-collapse-item rounded>
      <mdui-list-item slot="header" icon="settings_applications" rounded @click="getSetting()">
        Anthropic Claude
        <mdui-icon slot="end-icon" name="keyboard_arrow_down"></mdui-icon>
      </mdui-list-item>
      <mdui-list-item nonclickable>
        <mdui-card variant="outlined">
          <p>这里设置Anthropic Claude的后端API，可用于文字生成，总结等</p>
          <mdui-text-field variant="outlined" label="API Key" type="password" toggle-password
                            :value="settings?.api_key || ''"
                            @change="settings.api_key = $event.target.value"></mdui-text-field>
          <mdui-text-field variant="outlined" label="API地址"
                            :value="settings?.endpoint || 'https://api.anthropic.com/v1/messages'"
                            @change="settings.endpoint = $event.target.value"></mdui-text-field>
          <p>默认文字生成模型</p>
          <mdui-radio-group :value="settings?.text_model || 'claude-3-5-sonnet-20241022'" style="margin-top: 0"
                            @change="settings.text_model=$event.target.value">
            <mdui-radio value="claude-3-5-sonnet-20241022">Claude 3.5 Sonnet</mdui-radio>
            <mdui-radio value="claude-3-5-haiku-20241022">Claude 3.5 Haiku</mdui-radio>
            <mdui-radio value="claude-3-opus-20240229">Claude 3 Opus</mdui-radio>
            <mdui-radio value="claude-3-sonnet-20240229">Claude 3 Sonnet</mdui-radio>
            <mdui-radio value="claude-3-haiku-20240307">Claude 3 Haiku</mdui-radio>
          </mdui-radio-group>
          <p>默认总结模型</p>
          <mdui-radio-group :value="settings?.summary_model || 'claude-3-haiku-20240307'" style="margin-top: 0"
                            @change="settings.summary_model=$event.target.value">
            <mdui-radio value="claude-3-5-sonnet-20241022">Claude 3.5 Sonnet</mdui-radio>
            <mdui-radio value="claude-3-5-haiku-20241022">Claude 3.5 Haiku</mdui-radio>
            <mdui-radio value="claude-3-haiku-20240307">Claude 3 Haiku</mdui-radio>
          </mdui-radio-group>
          <p>最大Token数</p>
          <mdui-text-field variant="outlined" label="Max Tokens"
                            :value="settings?.max_tokens || 4096"
                            @change="settings.max_tokens = parseInt($event.target.value) || 4096"
                            type="number"></mdui-text-field>
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
