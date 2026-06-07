import React from 'react';
import { Tabs } from 'expo-router';
import { Ionicons } from '@expo/vector-icons';
import { View } from 'react-native';
import { t } from '../../i18n';

export default function TabLayout() {
  return (
    <Tabs
      screenOptions={{
        headerShown: false,
        tabBarActiveTintColor: '#007AFF',
        tabBarInactiveTintColor: '#8E8E93',
        tabBarStyle: { backgroundColor: '#FFFFFF', borderTopColor: '#C6C6C8' },
      }}
    >
      <Tabs.Screen
        name="index"
        options={{
          tabBarLabel: t('tabs.home'),
          tabBarIcon: ({ color, focused }) => (
            <Ionicons name={focused ? 'home' : 'home-outline'} size={24} color={color} />
          ),
        }}
      />
      <Tabs.Screen
        name="list"
        options={{
          tabBarLabel: t('tabs.list'),
          tabBarIcon: ({ color, focused }) => (
            <Ionicons name={focused ? 'grid' : 'grid-outline'} size={24} color={color} />
          ),
        }}
      />
      <Tabs.Screen
        name="community"
        options={{
          tabBarLabel: t('tabs.community'),
          tabBarIcon: ({ focused }) => (
            <View
              style={{
                width: 48, height: 48, borderRadius: 24,
                backgroundColor: '#007AFF',
                alignItems: 'center', justifyContent: 'center',
                marginBottom: 8,
                shadowColor: '#007AFF', shadowOffset: { width: 0, height: 4 },
                shadowOpacity: 0.4, shadowRadius: 8, elevation: 8,
              }}
            >
              <Ionicons name="add" size={28} color="white" />
            </View>
          ),
        }}
      />
      <Tabs.Screen
        name="history"
        options={{
          tabBarLabel: t('tabs.history'),
          tabBarIcon: ({ color, focused }) => (
            <Ionicons name={focused ? 'receipt' : 'receipt-outline'} size={24} color={color} />
          ),
        }}
      />
      <Tabs.Screen
        name="profile"
        options={{
          tabBarLabel: t('tabs.profile'),
          tabBarIcon: ({ color, focused }) => (
            <Ionicons name={focused ? 'person' : 'person-outline'} size={24} color={color} />
          ),
        }}
      />
    </Tabs>
  );
}
