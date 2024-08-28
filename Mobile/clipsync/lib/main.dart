import 'package:clipsync/home_page.dart';
import 'package:flutter/material.dart';
import 'package:flutter_foreground_task/flutter_foreground_task.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  // Initialize Foreground Task
  FlutterForegroundTask.init(
    androidNotificationOptions: AndroidNotificationOptions(
      channelId: 'clip_sync_channel',
      channelName: 'ClipSync Background Service',
      channelDescription:
      'This notification keeps the ClipSync service running in the background.',
      channelImportance: NotificationChannelImportance.LOW,
      priority: NotificationPriority.LOW,
      showBadge: true,
    ),
    iosNotificationOptions:
    const IOSNotificationOptions(showNotification: true, playSound: false),
    foregroundTaskOptions: const ForegroundTaskOptions(
      autoRunOnBoot: true,
    ),
  );

  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'ClipSync',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(
          seedColor: const Color(0xFF05245b), // Primary color
          brightness: Brightness.dark,
        ),
        useMaterial3: true,
        appBarTheme: const AppBarTheme(
          backgroundColor: Colors.black87,
          titleTextStyle: TextStyle(color: Colors.white, fontSize: 20),
        ),
        elevatedButtonTheme: ElevatedButtonThemeData(
          style: ElevatedButton.styleFrom(
            backgroundColor: const Color(0xFFefb32e), // Secondary color
            foregroundColor: Colors.black,
          ),
        ),
        floatingActionButtonTheme: const FloatingActionButtonThemeData(
          backgroundColor: Color(0xFFefb32e)
        ),
        textTheme: const TextTheme(
          bodyLarge: TextStyle(color: Colors.white),
          bodyMedium: TextStyle(color: Colors.white70),
        ),
        inputDecorationTheme: InputDecorationTheme(
          filled: true,
          fillColor: Colors.grey[800],
          border: OutlineInputBorder(
            borderSide: BorderSide(color: Colors.grey[700]!),
            borderRadius: BorderRadius.circular(12.0),
          ),
          enabledBorder: OutlineInputBorder(
            borderSide: BorderSide(color: Colors.grey[700]!),
            borderRadius: BorderRadius.circular(12.0),
          ),
          focusedBorder: OutlineInputBorder(
            borderSide: const BorderSide(color: Color(0xFFefb32e)),
            borderRadius: BorderRadius.circular(12.0),
          ),
          errorBorder: OutlineInputBorder(
            borderSide: const BorderSide(color: Colors.redAccent),
            borderRadius: BorderRadius.circular(12.0),
          ),
          focusedErrorBorder: OutlineInputBorder(
            borderSide: const BorderSide(color: Colors.redAccent),
            borderRadius: BorderRadius.circular(12.0),
          ),
        ),
      ),
      themeMode: ThemeMode.system, // Use system theme
      home: const MyHomePage(title: "ClipSync"),
    );
  }
}