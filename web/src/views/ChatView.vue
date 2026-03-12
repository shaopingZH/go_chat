<template>
  <main class="tg-layout relative overflow-hidden">
    <!-- Sidebar -->
    <!-- Sidebar -->
    <aside class="tg-sidebar relative overflow-hidden" :class="{ 'tg-sidebar--hidden': isMobile && showMobileChat }">
      <div class="h-full w-full relative bg-[var(--tg-bg-sidebar)]">
        <Transition :name="sidebarTransitionName">
          <!-- Chat List Panel (Level 0) -->
          <div v-if="!chat.showLeftSidebar" key="list" class="flex flex-col h-full w-full absolute inset-0">
            <div class="tg-header">
              <button class="tg-icon-btn -ml-2" @click="chat.openLeftSidebar('profile')">
                <div v-if="auth.user?.avatar" class="w-8 h-8 rounded-full overflow-hidden border border-[var(--tg-border)]">
                  <img :src="auth.user.avatar" class="w-full h-full object-cover">
                </div>
                <IconMenu v-else />
              </button>
              <div class="tg-search-wrapper">
                <IconSearch class="text-[var(--tg-text-gray)] w-5 h-5 flex-shrink-0" />
                <input 
                  v-model.trim="searchKeyword" 
                  class="tg-search-input" 
                  type="text" 
                  placeholder="Search" 
                />
              </div>
              <button class="tg-icon-btn" @click="toggleTheme($event)">
                <IconSun v-if="isDark" class="w-5 h-5" />
                <IconMoon v-else class="w-5 h-5" />
              </button>

              <button class="tg-icon-btn -mr-2 relative" @click="chat.openLeftSidebar('friends')">
                <IconUser class="w-5 h-5" />
                <span v-if="pendingRequestsCount > 0" class="absolute top-1 right-1 w-2.5 h-2.5 bg-red-500 rounded-full border-2 border-[var(--tg-bg-sidebar)]"></span>
              </button>
            </div>
            
            <div class="flex justify-between items-center px-4 py-2 border-b border-[var(--tg-border)]">
              <span class="text-sm font-bold text-[var(--tg-text-gray)]">Chats</span>
              <button class="w-6 h-6 flex items-center justify-center rounded-full bg-[var(--tg-primary)] text-white hover:opacity-90 active:scale-95 transition-all shadow-md shadow-[var(--tg-primary)]/20" @click="chat.openLeftSidebar('groupOptions')">
                <span class="text-lg leading-none mb-0.5">+</span>
              </button>
            </div>

            <div class="tg-list">
              <div v-if="chat.isLoadingConversations" class="flex flex-col items-center justify-center mt-10 gap-3 opacity-50">
                <div class="w-6 h-6 border-2 border-[var(--tg-text-blue)] border-t-transparent rounded-full animate-spin"></div>
                <span class="text-sm text-[var(--tg-text-gray)]">Loading chats...</span>
              </div>
              <template v-else>
                <ConversationItem
                  v-for="item in filteredConversations"
                  :key="item.key"
                  :target-id="item.targetId"
                  :title="item.chatType === 'group' ? item.title : item.title"
                  :subtitle="item.lastText"
                  :unread="item.unread"
                  :active="item.key === chat.activeKey"
                  :last-at="item.lastAt"
                  :chat-type="item.chatType"
                  :avatar="item.avatar"
                  @select="selectConversation(item.key)"
                  @contextmenu="(e) => openContextMenu(e, item.key)"
                />
                <div v-if="filteredConversations.length === 0" class="text-center text-[var(--tg-text-gray)] text-sm mt-10">
                  No chats found.
                </div>
              </template>
            </div>
          </div>

          <!-- Detail Panels (Level 1+) -->
          <MyProfileSidebar v-else-if="chat.leftSidebarType === 'profile'" key="profile" class="absolute inset-0" />
          <FriendsSidebar v-else-if="chat.leftSidebarType === 'friends'" key="friends" class="absolute inset-0" />
          <AddFriendSidebar v-else-if="chat.leftSidebarType === 'addFriend'" key="addFriend" class="absolute inset-0" />
          <GroupOptionsSidebar v-else-if="chat.leftSidebarType === 'groupOptions'" key="groupOptions" class="absolute inset-0" />
          <CreateGroupSidebar v-else-if="chat.leftSidebarType === 'createGroup'" key="createGroup" class="absolute inset-0" />
          <JoinGroupSidebar v-else-if="chat.leftSidebarType === 'joinGroup'" key="joinGroup" class="absolute inset-0" />
        </Transition>
      </div>
    </aside>

    <!-- Chat Area -->
    <section class="tg-chat" :class="{ 'tg-chat--hidden': isMobile && !showMobileChat }">
      <header class="tg-chat-header relative flex items-center px-4 overflow-hidden">
        <!-- Back Button (Mobile) -->
        <button v-if="isMobile" class="tg-icon-btn -ml-2 mr-2 flex-shrink-0" @click.stop="backToList">
          <IconBack />
        </button>

        <!-- Avatar (Always visible, acts as close search in search mode) -->
        <div v-if="chat.activeConversation" class="w-[42px] h-[42px] rounded-full overflow-hidden flex-shrink-0 border border-[var(--tg-border)] flex items-center justify-center bg-[var(--tg-bg-hover)] mr-3 cursor-pointer z-20 transition-transform active:scale-95" @click="chat.isSearching ? chat.closeSearch() : viewHeaderProfile()">
           <img v-if="headerAvatar" :src="headerAvatar" class="w-full h-full object-cover">
           <span v-else class="text-[15px] font-bold text-[var(--tg-text-gray)] uppercase">
             {{ chat.activeConversation ? (chat.activeConversation.title || "?")[0] : "?" }}
           </span>
        </div>

        <div class="flex-1 relative h-full min-w-0" v-if="chat.activeConversation">
          <!-- Normal Header Content -->
          <Transition name="header-fade">
            <div v-if="!chat.isSearching" class="absolute inset-0 flex items-center justify-between w-full h-full origin-left">
              <div class="flex flex-col cursor-pointer flex-1 min-w-0" @click="viewHeaderProfile">
                <h3 class="text-[16px] font-bold text-[var(--tg-text-white)] leading-tight truncate">
                  {{ activeTitle }}
                </h3>
                <span class="text-[14px] text-[var(--tg-text-gray)] leading-tight truncate">
                  {{ statusText }}
                </span>
              </div>
              
              <!-- Search Toggle -->
              <button class="tg-icon-btn -mr-2" @click.stop="chat.openSearch()">
                <IconSearch class="w-5 h-5" />
              </button>
            </div>
          </Transition>

          <!-- Search Header Content (The Pill) -->
          <Transition name="search-expand">
            <div v-if="chat.isSearching" class="absolute inset-0 flex items-center w-full h-full origin-right">
              <div class="flex-1 flex items-center tg-search-pill h-[42px] rounded-full pl-4 pr-2 transition-all">
                <!-- Loader or Search Icon -->
                <div class="w-5 h-5 mr-3 flex-shrink-0 flex items-center justify-center text-[var(--tg-text-gray)]">
                  <svg v-if="chat.isSearchingAPI || (chat.searchKeyword && chat.searchResults.length === 0)" class="animate-spin w-5 h-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2.5"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V1.5C5.373 1.5 0 6.873 0 12h4z"></path>
                  </svg>
                  <IconSearch v-else class="w-5 h-5" />
                </div>

                <!-- Input Area -->
                <input
                  ref="searchInput"
                  v-model="chat.searchKeyword"
                  class="bg-transparent text-[var(--tg-text-white)] text-[16px] focus:outline-none flex-1 placeholder:text-[var(--tg-text-hint)] min-w-0 font-medium"
                  type="text"
                  placeholder="Search"
                  @keydown.enter="chat.nextSearchResult"
                />

                <!-- Right side controls (Up, Down, Close) -->
                <div class="flex items-center gap-1 pl-3 shrink-0 text-[var(--tg-text-gray)]">
                  <span v-if="chat.searchResults.length > 0" class="text-[14px] font-medium whitespace-nowrap mr-2">
                    {{ chat.searchResults.length - chat.searchCurrentIndex }} of {{ chat.searchResults.length }}
                  </span>

                  <div class="flex items-center" v-if="chat.searchResults.length > 0">
                    <button
                      class="w-8 h-8 flex items-center justify-center rounded-full hover:bg-[var(--tg-bg-hover)] transition-colors active:scale-95"
                      :class="{ 'opacity-30 cursor-not-allowed': chat.searchResults.length === 0 || chat.searchCurrentIndex === chat.searchResults.length - 1 }"
                      :disabled="chat.searchResults.length === 0 || chat.searchCurrentIndex === chat.searchResults.length - 1"
                      @click="chat.nextSearchResult"
                    >
                      <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m18 15-6-6-6 6"/></svg>
                    </button>
                    <button
                      class="w-8 h-8 flex items-center justify-center rounded-full hover:bg-[var(--tg-bg-hover)] transition-colors active:scale-95 mr-1"
                      :class="{ 'opacity-30 cursor-not-allowed': chat.searchResults.length === 0 || chat.searchCurrentIndex === 0 }"
                      :disabled="chat.searchResults.length === 0 || chat.searchCurrentIndex === 0"
                      @click="chat.prevSearchResult"
                    >
                      <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m6 9 6 6 6-6"/></svg>
                    </button>
                  </div>

                  <button
                    class="w-8 h-8 flex items-center justify-center rounded-full hover:bg-[var(--tg-bg-hover)] transition-colors active:scale-95 text-[var(--tg-text-gray)]"
                    @click="chat.closeSearch()"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
                  </button>
                </div>
              </div>
            </div>
          </Transition>
        </div>
      </header>

      <div class="relative flex-1 min-h-0 flex flex-col">
        <div class="tg-chat-content" ref="scrollContainer" @scroll="handleScroll">
          <!-- Loader when fetching history -->
          <div v-if="isLoadingHistory" class="w-full flex justify-center py-4">
            <svg class="animate-spin w-6 h-6 text-[var(--tg-text-gray)]" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2.5"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V1.5C5.373 1.5 0 6.873 0 12h4z"></path>
            </svg>
          </div>

          <div v-if="!chat.activeConversation" class="h-full flex flex-col items-center justify-center opacity-50">
             <div class="bg-[var(--tg-bg-hover)] p-4 rounded-full mb-4">
               <span class="text-4xl">👋</span>
             </div>
             <p class="text-sm text-[var(--tg-text-gray)]">Select a chat to start messaging</p>
          </div>
          
          <!-- Chat Stream Container -->
          <div v-else class="w-full max-w-[720px] mx-auto flex flex-col min-h-full px-4 pt-4 pb-4">
            <div v-for="group in activeMessagesByDate" :key="group.date" class="flex flex-col relative">
              <!-- Date Divider -->
              <div class="flex justify-center my-4 sticky top-2 z-10">
                <span class="bg-[var(--tg-bg-date)] text-[var(--tg-text-white)] text-[14px] px-3 py-1 rounded-full font-semibold shadow-sm backdrop-blur-sm tracking-tight">
                  {{ group.dateLabel }}
                </span>
              </div>

              <MessageBubble
                v-for="message in group.messages"
                :key="message.id"
                :message="message"
                :mine="isMine(message)"
                @preview="handleImagePreview"
                @imageLoaded="handleImageLoad"
              />
            </div>
          </div>
        </div>

        <!-- Scroll to bottom FAB -->
        <Transition name="fade">
          <button 
            v-if="showScrollBottomBtn" 
            @click="scrollToBottom" 
            class="absolute bottom-4 right-4 md:right-8 w-11 h-11 bg-[var(--tg-bg-header)] rounded-full flex items-center justify-center shadow-lg border border-[var(--tg-border)] hover:bg-[var(--tg-bg-hover)] transition-colors z-20 group"
          >
            <svg class="w-6 h-6 text-[var(--tg-text-gray)] group-hover:text-[var(--tg-text-white)]" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 14l-7 7m0 0l-7-7m7 7V3"></path></svg>
            <span v-if="unreadMessagesBelow > 0" class="absolute -top-1 -right-1 bg-[var(--tg-accent)] text-white text-[11px] font-bold px-1.5 min-w-[20px] h-[20px] rounded-full flex items-center justify-center border-2 border-[var(--tg-bg-body)]">
              {{ unreadMessagesBelow > 99 ? '99+' : unreadMessagesBelow }}
            </span>
          </button>
        </Transition>
      </div>

      <!-- Input Area -->
      <div class="tg-input-area" v-if="chat.activeConversation">
        <template v-if="chatRestrictionReason">
          <div class="w-full flex justify-center py-3">
            <span class="text-[14px] text-[var(--tg-text-gray)] bg-[var(--tg-bg-hover)] px-4 py-2 rounded-full">{{ chatRestrictionReason }}</span>
          </div>
        </template>
        <template v-else>
          <!-- Image Preview Floating Card -->
          <Transition name="fade">
            <div v-if="imagePreviewUrl" class="w-full max-w-[720px] mx-auto px-1 mb-2">
              <div class="relative inline-block group">
                <div class="w-20 h-20 rounded-xl overflow-hidden border border-[var(--tg-border)] shadow-sm bg-[var(--tg-bg-hover)]">
                  <img :src="imagePreviewUrl" alt="preview" class="w-full h-full object-cover">
                </div>
                <button 
                  class="absolute -top-2 -right-2 w-6 h-6 bg-[var(--tg-bg-header)] text-[var(--tg-text-gray)] hover:text-[var(--tg-text-white)] rounded-full flex items-center justify-center shadow-md border border-[var(--tg-border)] transition-colors opacity-100 md:opacity-0 md:group-hover:opacity-100" 
                  @click="clearSelectedImage"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
                </button>
              </div>
            </div>
          </Transition>

          <div class="tg-input-container relative">
              <!-- Emoji Picker -->
              <Transition name="emoji-pop">
                <div v-if="showEmojiPicker" ref="emojiPickerRef" class="absolute bottom-full left-0 mb-3 z-50">
                  <EmojiPicker 
                    :theme="isDark ? 'dark' : 'light'" 
                    :native="true"
                    :hide-search="true"
                    :hide-group-names="true"
                    @select="onSelectEmoji" 
                  />
                </div>
              </Transition>

              <input
                ref="imageInput"
                class="hidden"
                type="file"
                accept="image/png,image/jpeg,image/gif,image/webp"
                @change="onImageSelected"
              >

              <button class="tg-send-btn flex-shrink-0" type="button" @click="triggerImagePicker">
                <span class="text-xl leading-none">+</span>
              </button>

              <!-- Emoji Toggle Button -->
              <button 
                class="tg-send-btn emoji-toggle-btn flex-shrink-0" 
                :class="{ 'text-[var(--tg-text-blue)]': showEmojiPicker }"
                type="button" 
                @click.stop="toggleEmojiPicker"
              >
                <svg xmlns="http://www.w3.org/2000/svg" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M8 14s1.5 2 4 2 4-2 4-2"/><line x1="9" y1="9" x2="9.01" y2="9"/><line x1="15" y1="9" x2="15.01" y2="9"/></svg>
              </button>
              
              <textarea
                ref="messageInput"
                v-model="draft"
                class="tg-input-field"
                placeholder="Message"
                rows="1"
                @keydown.enter.prevent="sendMessage"
              ></textarea>
              
              <button class="tg-send-btn" :class="{ 'opacity-50 cursor-not-allowed': !canSend || sendingImage }" :disabled="!canSend || sendingImage" @click="sendMessage">
                <span v-if="sendingImage" class="text-[11px]">...</span>
                <IconSend v-else />
              </button>
           </div>
           <div v-if="displayError" class="text-center text-red-400 text-xs mt-2">{{ displayError }}</div>
        </template>
      </div>
    </section>

    <!-- Right Sidebar -->
    <transition name="slide-right">
      <div v-if="chat.showRightSidebar" class="tg-right-sidebar-wrapper">
        <aside class="tg-right-sidebar-inner bg-[var(--tg-bg-sidebar)] h-full shadow-[-4px_0_15px_rgba(0,0,0,0.2)] md:shadow-none border-l border-[var(--tg-border)]">
          <PublicProfileSidebar v-if="chat.rightSidebarType === 'publicProfile'" />
          <GroupProfileSidebar v-else-if="chat.rightSidebarType === 'groupProfile'" />
        </aside>
      </div>
    </transition>

    <!-- Fullscreen Image Preview Overlay -->
    <Transition name="fade">
      <div v-if="previewImageUrl" class="fixed inset-0 z-[10000] bg-black/90 backdrop-blur-sm flex items-center justify-center" @click="closeImagePreview">
        <button class="absolute top-4 right-4 w-10 h-10 flex items-center justify-center bg-white/10 hover:bg-white/20 rounded-full text-white transition-colors" @click="closeImagePreview">
          <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
        </button>
        <img :src="previewImageUrl" class="max-w-[95vw] max-h-[95vh] object-contain select-none" @click.stop>
      </div>
    </Transition>

    <!-- Context Menu Overlay -->
    <Transition name="fade">
      <div v-if="contextMenu.show" class="fixed inset-0 z-[9999]" @click.stop="closeContextMenu" @contextmenu.prevent="closeContextMenu">
        <div 
          class="absolute bg-[var(--tg-bg-menu)] shadow-[0_4px_24px_rgba(0,0,0,0.15)] rounded-[14px] p-1 w-48 flex flex-col z-[10000] tg-menu-pop"
          :style="{ top: `${contextMenu.y}px`, left: `${contextMenu.x}px`, '--menu-transform-origin': contextMenu.transformOrigin }"
          @click.stop
        >
          <button 
            class="flex items-center px-3 py-2 text-[14px] text-red-500 hover:bg-[var(--tg-bg-hover)] rounded-lg transition-colors w-full text-left font-medium"
            @click="deleteFromContext"
          >
            <IconDelete class="w-5 h-5 mr-3 text-red-500" />
            Delete Chat
          </button>
        </div>
      </div>
    </Transition>

    <!-- Global Toasts Container -->
    <div class="fixed top-4 right-4 md:top-6 md:right-6 z-[10000] flex flex-col gap-3 pointer-events-none items-end">
      <TransitionGroup name="toast-slide">
        <div 
          v-for="toast in chat.toasts" 
          :key="toast.id" 
          class="bg-[var(--tg-bg-menu)] backdrop-blur-xl border border-[var(--tg-border)] shadow-[0_8px_30px_rgba(0,0,0,0.2)] rounded-2xl p-4 text-[var(--tg-text-white)] flex items-start gap-3 w-80 max-w-[90vw] pointer-events-auto transition-all"
        >
          <!-- Optional Icon based on type -->
          <div v-if="toast.type === 'success'" class="text-green-400 mt-0.5 flex-shrink-0">
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path><polyline points="22 4 12 14.01 9 11.01"></polyline></svg>
          </div>
          <div v-else-if="toast.type === 'error'" class="text-red-400 mt-0.5 flex-shrink-0">
             <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
          </div>
          <div v-else class="text-[var(--tg-primary)] mt-0.5 flex-shrink-0">
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="16" x2="12" y2="12"></line><line x1="12" y1="8" x2="12.01" y2="8"></line></svg>
          </div>

          <div class="flex-1 flex flex-col min-w-0 pr-2">
            <div v-if="toast.title" class="font-bold text-[15px] mb-1 truncate">{{ toast.title }}</div>
            <div class="text-[14px] leading-snug text-[var(--tg-text-gray)]">
              {{ toast.message }}
            </div>
          </div>
          <button class="w-6 h-6 rounded-full flex items-center justify-center bg-transparent hover:bg-[var(--tg-bg-hover)] text-[var(--tg-text-gray)] transition-colors active:scale-95 flex-shrink-0 -mr-1" @click="chat.removeToast(toast.id)">
             <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
          </button>
        </div>
      </TransitionGroup>
    </div>
  </main>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from "vue"
import { useRouter } from "vue-router"
import ConversationItem from "../components/ConversationItem.vue"
import MyProfileSidebar from "../components/MyProfileSidebar.vue"
import FriendsSidebar from "../components/FriendsSidebar.vue"
import AddFriendSidebar from "../components/AddFriendSidebar.vue"
import CreateGroupSidebar from "../components/CreateGroupSidebar.vue"
import JoinGroupSidebar from "../components/JoinGroupSidebar.vue"
import GroupOptionsSidebar from "../components/GroupOptionsSidebar.vue"
import MessageBubble from "../components/MessageBubble.vue"
import IconMenu from "../components/icons/IconMenu.vue"
import IconSearch from "../components/icons/IconSearch.vue"
import IconSend from "../components/icons/IconSend.vue"
import IconBack from "../components/icons/IconBack.vue"
import IconSun from "../components/icons/IconSun.vue"
import IconMoon from "../components/icons/IconMoon.vue"
import IconUser from "../components/icons/IconUser.vue"
import IconDelete from "../components/icons/IconDelete.vue"
import PublicProfileSidebar from "../components/PublicProfileSidebar.vue"
import GroupProfileSidebar from "../components/GroupProfileSidebar.vue"

import { useAuthStore } from "../stores/auth"
import { useChatStore } from "../stores/chat"
import { useFriendStore } from "../stores/friend"
import { useTheme } from "../composables/useTheme"
import type { ChatMessage } from "../types/chat"

// Emoji Picker
import EmojiPicker from 'vue3-emoji-picker'
import 'vue3-emoji-picker/css'

const router = useRouter()
const auth = useAuthStore()
const chat = useChatStore()
const friendStore = useFriendStore()
const { isDark, toggleTheme } = useTheme()

const pendingRequestsCount = computed(() => {
  return friendStore.requests.filter(r => r.status === 'pending').length
})

// Context Menu State
const contextMenu = ref({
  show: false,
  x: 0,
  y: 0,
  key: "",
  transformOrigin: "top left"
})

function openContextMenu(event: MouseEvent, key: string) {
  contextMenu.value.show = true
  contextMenu.value.key = key
  
  // Adjust position so it doesn't go off-screen
  let x = event.clientX
  let y = event.clientY
  let originX = "left"
  let originY = "top"
  
  const menuWidth = 192 // w-48
  const menuHeight = 50 // Approx height
  
  if (x + menuWidth > window.innerWidth) {
    x -= menuWidth
    originX = "right"
  }
  if (y + menuHeight > window.innerHeight) {
    y -= menuHeight
    originY = "bottom"
  }
  
  contextMenu.value.x = x
  contextMenu.value.y = y
  contextMenu.value.transformOrigin = `${originY} ${originX}`
}

function closeContextMenu() {
  contextMenu.value.show = false
}

function deleteFromContext() {
  if (contextMenu.value.key) {
    chat.removeConversation(contextMenu.value.key)
  }
  closeContextMenu()
}

const draft = ref("")
const showEmojiPicker = ref(false)
const emojiPickerRef = ref<HTMLElement | null>(null)
const messageInput = ref<HTMLTextAreaElement | null>(null)

function toggleEmojiPicker() {
  showEmojiPicker.value = !showEmojiPicker.value
}

function onSelectEmoji(emoji: any) {
  if (!messageInput.value) return
  
  const textarea = messageInput.value
  const start = textarea.selectionStart
  const end = textarea.selectionEnd
  const text = draft.value
  
  draft.value = text.substring(0, start) + emoji.i + text.substring(end)
  
  // Restore focus and cursor position after DOM update
  nextTick(() => {
    textarea.focus()
    const newCursorPos = start + emoji.i.length
    textarea.setSelectionRange(newCursorPos, newCursorPos)
  })
}

// Close picker when clicking outside
function handleClickOutside(e: MouseEvent) {
  if (showEmojiPicker.value && emojiPickerRef.value && !emojiPickerRef.value.contains(e.target as Node)) {
    const btn = document.querySelector('.emoji-toggle-btn')
    if (btn && btn.contains(e.target as Node)) return
    showEmojiPicker.value = false
  }
}
const searchKeyword = ref("")
const localError = ref("")
const scrollContainer = ref<HTMLDivElement | null>(null)
const imageInput = ref<HTMLInputElement | null>(null)
const imagePreviewUrl = ref("")
const pendingImage = ref<File | null>(null)
const pendingImageName = ref("")
const sendingImage = ref(false)

const previewImageUrl = ref("")
const isLoadingHistory = ref(false)

const showScrollBottomBtn = ref(false)
const unreadMessagesBelow = ref(0)
const scrollPositions = ref<Record<string, number>>({})
const unreadMessagesBelowMap = ref<Record<string, number>>({})

function scrollToBottom() {
  if (!scrollContainer.value) return
  scrollContainer.value.scrollTo({
    top: scrollContainer.value.scrollHeight,
    behavior: 'smooth'
  })
  unreadMessagesBelow.value = 0
  showScrollBottomBtn.value = false
  if (chat.activeKey) {
    unreadMessagesBelowMap.value[chat.activeKey] = 0
  }
}

async function handleScroll(e: Event) {
  const target = e.target as HTMLElement
  if (!target || !chat.activeKey) return
  
  // Save current scroll position for this chat
  scrollPositions.value[chat.activeKey] = target.scrollTop

  const isNearBottom = target.scrollHeight - target.scrollTop - target.clientHeight < 200
  showScrollBottomBtn.value = !isNearBottom
  if (isNearBottom) {
    unreadMessagesBelow.value = 0
    unreadMessagesBelowMap.value[chat.activeKey] = 0
  }

  // If we scroll to the very top (or close to it) and we are not already loading
  if (target.scrollTop <= 100 && !isLoadingHistory.value) {
    // Use the explicit hasMoreByKey state from store
    if (chat.hasMoreByKey[chat.activeKey] !== false) {
      isLoadingHistory.value = true
      
      // Save current scroll height to restore scroll position after DOM update
      const previousScrollHeight = target.scrollHeight
      
      const loaded = await chat.loadMoreHistory(chat.activeKey)
      
      if (loaded) {
        await nextTick()
        // Restore scroll position so the user stays exactly where they were looking
        const newScrollHeight = target.scrollHeight
        target.scrollTop = newScrollHeight - previousScrollHeight
      }
      
      isLoadingHistory.value = false
    }
  }
}

function handleImagePreview(url: string) {
  previewImageUrl.value = url
}

function handleImageLoad() {
  if (!scrollContainer.value) return
  // If user is reasonably close to the bottom, auto-scroll down to accommodate the newly loaded image
  const isNearBottom = scrollContainer.value.scrollHeight - scrollContainer.value.scrollTop - scrollContainer.value.clientHeight < 500
  if (isNearBottom) {
    scrollContainer.value.scrollTop = scrollContainer.value.scrollHeight
  }
}

function closeImagePreview() {
  previewImageUrl.value = ""
}

const maxImageSizeBytes = 5 * 1024 * 1024

const activeMessagesByDate = computed(() => {
  const groups: { date: string; dateLabel: string; messages: ChatMessage[] }[] = []
  if (!chat.activeMessages || chat.activeMessages.length === 0) return groups

  let currentGroup = {
    date: "",
    dateLabel: "",
    messages: [] as ChatMessage[]
  }

  chat.activeMessages.forEach((message) => {
    const msgDate = new Date(message.created_at)
    const dateKey = `${msgDate.getFullYear()}-${msgDate.getMonth() + 1}-${msgDate.getDate()}`

    if (currentGroup.date !== dateKey) {
      if (currentGroup.messages.length > 0) {
        groups.push(currentGroup)
      }
      currentGroup = {
        date: dateKey,
        dateLabel: formatDateDivider(message.created_at),
        messages: [message]
      }
    } else {
      currentGroup.messages.push(message)
    }
  })

  if (currentGroup.messages.length > 0) {
    groups.push(currentGroup)
  }

  return groups
})

const mobileMedia = window.matchMedia("(max-width: 768px)")
const isMobile = ref(mobileMedia.matches)
const showMobileChat = ref(!mobileMedia.matches)

const sidebarTransitionName = ref("push-sidebar")
const panelLevels: Record<string, number> = {
  "profile": 1,
  "friends": 1,
  "groupOptions": 1,
  "addFriend": 2,
  "createGroup": 2,
  "joinGroup": 2
}

watch(
  () => [chat.leftSidebarType, chat.showLeftSidebar],
  ([newType, showSidebar], [oldType, oldShowSidebar]) => {
    const newLevel = !showSidebar ? 0 : (panelLevels[newType as string] || 1)
    const oldLevel = !oldShowSidebar ? 0 : (panelLevels[oldType as string] || 1)
    
    if (newLevel > oldLevel) {
      sidebarTransitionName.value = "push-sidebar"
    } else if (newLevel < oldLevel) {
      sidebarTransitionName.value = "push-sidebar-reverse"
    } else {
      // Same level switch
      sidebarTransitionName.value = "sidebar-fade"
    }
  },
  { immediate: false }
)

const statusText = computed(() => {
  const current = chat.activeConversation
  if (!current) return ""

  if (current.chatType === "group") {
    const members = chat.groupMembersCache[current.targetId]
    if (!members) return "group"
    
    // Calculate online count from cache with explicit ID normalization
    const onlineCount = members.filter(m => {
      const uid = Number(m.user_id || m.id)
      const status = chat.userOnlineStatusCache[uid]
      return status?.online === true
    }).length

    if (onlineCount > 0) {
      return `${members.length} members, ${onlineCount} online`
    }
    return `${members.length} members`
  }

  // Private chat: show peer presence
  const peerStatus = chat.userOnlineStatusCache[current.targetId]
  if (peerStatus) {
    if (peerStatus.online) return "online"
    if (peerStatus.last_seen_at) return formatLastSeen(peerStatus.last_seen_at)
    return "last seen a long time ago"
  }

  if (chat.wsStatus === "connected") return "online"
  if (chat.wsStatus === "connecting") return "connecting..."
  return "waiting for network..."
})

function formatLastSeen(dateStr: string): string {
  if (!dateStr) return "last seen a long time ago"
  const date = new Date(dateStr)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffSec = Math.floor(diffMs / 1000)
  const diffMin = Math.floor(diffSec / 60)
  const diffHour = Math.floor(diffMin / 60)
  const diffDay = Math.floor(diffHour / 24)

  if (diffSec < 60) return "last seen just now"
  if (diffMin < 60) return `last seen ${diffMin} ${diffMin === 1 ? 'minute' : 'minutes'} ago`
  if (diffHour < 24) return `last seen ${diffHour} ${diffHour === 1 ? 'hour' : 'hours'} ago`
  if (diffDay === 1) return "last seen yesterday"
  if (diffDay < 7) return `last seen ${diffDay} days ago`
  
  return `last seen on ${date.toLocaleDateString()}`
}

const activeTitle = computed(() => {
  const current = chat.activeConversation
  if (!current) return "Go-Chat"
  return current.title
})

const chatRestrictionReason = computed(() => {
  if (chat.errorMessage === "not friend") {
    return "您已不是对方的好友，无法发送消息"
  }
  if (chat.errorMessage === "not group member") {
    return "您已不在该群组，无法发送消息"
  }
  return ""
})

const displayError = computed(() => {
  const err = localError.value || chat.errorMessage
  if (err === "not friend" || err === "not group member") return ""
  return err
})
const canSend = computed(() => Boolean(draft.value.trim()) || Boolean(pendingImage.value))

const filteredConversations = computed(() => {
  const keyword = searchKeyword.value.trim().toLowerCase()
  if (!keyword) return chat.conversations
  return chat.conversations.filter((item) => {
    return item.title.toLowerCase().includes(keyword) || item.lastText.toLowerCase().includes(keyword)
  })
})

const headerAvatar = computed(() => {
  const current = chat.activeConversation
  if (!current) return ""
  if (current.chatType === 'private') {
    const profile = chat.userProfileCache[current.targetId]
    return profile?.avatar || current.avatar
  }
  return current.avatar
})

function updateResponsiveState() {
  isMobile.value = mobileMedia.matches
  if (!isMobile.value) {
    showMobileChat.value = true
    return
  }
  showMobileChat.value = Boolean(chat.activeKey)
}

onMounted(async () => {
  mobileMedia.addEventListener("change", updateResponsiveState)
  window.addEventListener('click', handleClickOutside)
  
  try {
    localError.value = ""
    // 1. Ensure user/token is ready
    if (!auth.token) {
      router.push('/login')
      return
    }
    
    // Connect WebSocket immediately
    chat.connect()

    // Try fetch profile and conversations, but don't let it block the entire UI if it partially fails
    void auth.fetchMe().catch(console.error)
    void friendStore.loadRequests().catch(console.error)
    await chat.bootstrapConversations().catch(console.error)
    
    updateResponsiveState()
    
    // Initial fetch for presence if a conversation was auto-activated
    if (chat.activeConversation) {
      if (chat.activeConversation.chatType === 'group') {
        void chat.getGroupMembers(chat.activeConversation.targetId).catch(console.error)
        void chat.fetchGroupMembersPresence(chat.activeConversation.targetId).catch(console.error)
      } else {
        void chat.fetchUserPresence(chat.activeConversation.targetId).catch(console.error)
      }
    }
  } catch (error) {
    console.error("Initialization failed:", error)
    localError.value = error instanceof Error ? error.message : "Init failed"
    updateResponsiveState()
  }
})

onUnmounted(() => {
  mobileMedia.removeEventListener("change", updateResponsiveState)
  window.removeEventListener('click', handleClickOutside)
  clearSelectedImage()
  chat.disconnect()
})

watch(
  () => [chat.activeKey, chat.activeMessages.length],
  async (newVals, oldVals) => {
    if (isMobile.value && chat.activeKey) {
      showMobileChat.value = true
    }
    await nextTick()
    if (scrollContainer.value) {
      const newKey = newVals[0]
      const oldKey = oldVals[0]
      const newLen = newVals[1] as number
      const oldLen = oldVals[1] as number

      // If we just changed conversation (key changed), snap instantly
      if (newKey !== oldKey) {
        if (newKey && scrollPositions.value[newKey] !== undefined) {
          // Restore previously saved scroll position
          scrollContainer.value.scrollTop = scrollPositions.value[newKey]
          const isNearBottom = scrollContainer.value.scrollHeight - scrollPositions.value[newKey] - scrollContainer.value.clientHeight < 200
          showScrollBottomBtn.value = !isNearBottom
          unreadMessagesBelow.value = unreadMessagesBelowMap.value[newKey as string] || 0
        } else {
          // Default to bottom for new/unseen chats
          scrollContainer.value.scrollTop = scrollContainer.value.scrollHeight
          showScrollBottomBtn.value = false
          unreadMessagesBelow.value = 0
        }
      } else if (newLen > oldLen) {
        // If it's an initial load for this chat, snap instantly to bottom
        if (oldLen === 0) {
          scrollContainer.value.scrollTop = scrollContainer.value.scrollHeight
          showScrollBottomBtn.value = false
          unreadMessagesBelow.value = 0
        }
        // If just one new message arrived, and we are already near the bottom, smooth scroll
        else if (newLen - oldLen === 1) {
          const isNearBottom = scrollContainer.value.scrollHeight - scrollContainer.value.scrollTop - scrollContainer.value.clientHeight < 200
          if (isNearBottom || (chat.activeMessages[newLen - 1]?.sender?.id === auth.user?.id)) {
            scrollContainer.value.scrollTo({
              top: scrollContainer.value.scrollHeight,
              behavior: 'smooth'
            })
            showScrollBottomBtn.value = false
            unreadMessagesBelow.value = 0
            if (chat.activeKey) {
              unreadMessagesBelowMap.value[chat.activeKey] = 0
            }
          } else {
            // New message arrived while we are scrolled up
            unreadMessagesBelow.value += 1
            if (chat.activeKey) {
              unreadMessagesBelowMap.value[chat.activeKey] = unreadMessagesBelow.value
            }
          }
        }
      }
    }
  },
  { immediate: false }
)

watch(
  () => chat.activeKey,
  async (newKey) => {
    if (chat.showRightSidebar && chat.activeConversation) {
      const current = chat.activeConversation
      if (current.chatType === "group") {
        chat.openRightSidebar("groupProfile", current.targetId)
      } else if (current.chatType === "private") {
        chat.openRightSidebar("publicProfile", current.targetId)
      }
    }

    if (newKey && chat.activeConversation) {
      if (chat.activeConversation.chatType === "group") {
        await chat.getGroupMembers(chat.activeConversation.targetId)
        await chat.fetchGroupMembersPresence(chat.activeConversation.targetId)
      } else {
        await chat.fetchUserPresence(chat.activeConversation.targetId)
      }
    }
  }
)

async function selectConversation(key: string) {
  localError.value = ""
  try {
    await chat.activateConversation(key)
  } catch (error) {
    localError.value = error instanceof Error ? error.message : "Load failed"
  }
}

async function sendMessage() {
  const text = draft.value.trim()
  const file = pendingImage.value
  if (!text && !file) return

  localError.value = ""
  try {
    if (file) {
      sendingImage.value = true
      await chat.sendImage(file)
      clearSelectedImage()
    }
    if (text) {
      chat.sendMessage(text)
      draft.value = ""
    }
  } catch (error) {
    localError.value = error instanceof Error ? error.message : "Send failed"
  } finally {
    sendingImage.value = false
  }
}

const searchInput = ref<HTMLInputElement | null>(null)

let searchTimeout: ReturnType<typeof setTimeout>

watch(
  () => chat.searchKeyword,
  (newKeyword) => {
    clearTimeout(searchTimeout)
    if (!newKeyword.trim()) {
      chat.searchResults = []
      chat.searchCurrentIndex = 0
      return
    }
    searchTimeout = setTimeout(() => {
      chat.executeSearch()
    }, 300)
  }
)

watch(
  () => chat.isSearching,
  async (isSearching) => {
    if (isSearching) {
      await nextTick()
      searchInput.value?.focus()
    }
  }
)

watch(
  () => chat.searchCurrentIndex,
  async (index) => {
    if (chat.searchResults.length === 0 || index < 0) return
    const targetMessage = chat.searchResults[index]
    
    // Check if message is in current DOM
    let currentMsgs = chat.messagesByKey[chat.activeKey] || []
    let inDOM = currentMsgs.some((m: ChatMessage) => m.id === targetMessage.id)

    // If not in DOM, fetch history directly around the target message
    if (!inDOM && chat.activeKey) {
      chat.isSearchingAPI = true
      try {
        const loaded = await chat.loadHistoryAround(chat.activeKey, targetMessage.id)
        if (!loaded) {
          // Fallback: If API failed or didn't return it, inject manually
          chat.messagesByKey[chat.activeKey] = [targetMessage, ...currentMsgs]
        }
      } finally {
        chat.isSearchingAPI = false
      }
    }

    await nextTick()    
    // Find the element with data-message-id
    const el = document.querySelector(`[data-message-id="${targetMessage.id}"]`)
    if (el) {
      el.scrollIntoView({ behavior: 'smooth', block: 'center' })
      
      // Force a reflow to restart animation if it was already playing on THIS specific element
      void (el as HTMLElement).offsetWidth
      
      // Apply the Telegram style flash animation
      el.classList.add('tg-flash-highlight')
      
      setTimeout(() => {
        el.classList.remove('tg-flash-highlight')
      }, 1500)
    }
  }
)

function triggerImagePicker() {
  imageInput.value?.click()
}

function onImageSelected(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  target.value = ""
  if (!file) {
    return
  }

  if (!file.type.startsWith("image/")) {
    localError.value = "仅支持图片文件"
    return
  }
  if (file.size > maxImageSizeBytes) {
    localError.value = "图片不能超过 5MB"
    return
  }

  localError.value = ""
  clearSelectedImage()
  pendingImage.value = file
  pendingImageName.value = file.name
  imagePreviewUrl.value = URL.createObjectURL(file)
}

function clearSelectedImage() {
  if (imagePreviewUrl.value) {
    URL.revokeObjectURL(imagePreviewUrl.value)
  }
  imagePreviewUrl.value = ""
  pendingImage.value = null
  pendingImageName.value = ""
}

// We can remove shouldShowDateDivider if it's no longer used elsewhere
function isSameDay(date1: Date, date2: Date): boolean {
  return date1.getFullYear() === date2.getFullYear() &&
         date1.getMonth() === date2.getMonth() &&
         date1.getDate() === date2.getDate()
}

function formatDateDivider(dateStr: string): string {
  const date = new Date(dateStr)
  const today = new Date()
  const yesterday = new Date(today)
  yesterday.setDate(yesterday.getDate() - 1)

  if (isSameDay(date, today)) {
    return "今天"
  } else if (isSameDay(date, yesterday)) {
    return "昨天"
  } else {
    const options: Intl.DateTimeFormatOptions = { 
      month: 'long', 
      day: 'numeric' 
    }
    if (date.getFullYear() !== today.getFullYear()) {
      options.year = 'numeric'
    }
    return date.toLocaleDateString('zh-CN', options)
  }
}
function isMine(message: ChatMessage): boolean {
  return Number(message.sender?.id) === Number(auth.user?.id)
}

function backToList() {
  if (isMobile.value) {
    showMobileChat.value = false
    chat.activeKey = "" // Deselect
  }
}

function viewHeaderProfile() {
  const current = chat.activeConversation
  if (!current) return

  const targetType = current.chatType === 'private' ? 'publicProfile' : 'groupProfile'
  
  // Toggle logic: If already open with the same content, close it
  if (chat.showRightSidebar && chat.rightSidebarType === targetType && chat.rightSidebarPayload === current.targetId) {
    chat.closeRightSidebar()
  } else {
    chat.openRightSidebar(targetType, current.targetId)
  }
}
</script>

<style scoped>
/* Emoji Pop Animation */
.emoji-pop-enter-active,
.emoji-pop-leave-active {
  transition: all 0.2s cubic-bezier(0.33, 1, 0.68, 1);
}

.emoji-pop-enter-from,
.emoji-pop-leave-to {
  opacity: 0;
  transform: translateY(10px) scale(0.95);
}

/* Emoji Picker Theme Adaption */
:deep(.v3-emoji-picker) {
  --v3-bg: var(--tg-bg-menu);
  --v3-border-color: var(--tg-border);
  --v3-group-title-bg: var(--tg-bg-menu);
  --v3-group-title-color: var(--tg-text-blue);
  --v3-hover-bg: var(--tg-bg-hover);
  border: 1px solid var(--tg-border);
  backdrop-blur: 16px;
  width: 320px;
  height: 360px;
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.2);
  border-radius: 16px;
  overflow: hidden;
}

:deep(.v3-emoji-picker .v3-body) {
  padding: 8px;
}

/* Hide Search Bar */
:deep(.v3-emoji-picker .v3-search) {
  display: none;
}

/* Hide Footer (Hover Labels) */
:deep(.v3-emoji-picker .v3-footer) {
  display: none;
}

/* Style categories icons */
:deep(.v3-emoji-picker .v3-groups) {
  padding: 8px 12px;
  border-bottom: 1px solid var(--tg-border);
}
</style>
