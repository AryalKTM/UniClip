import 'package:shared_preferences/shared_preferences.dart';

class SavedAddress {
  final String name;
  final String ip;
  final String port;

  SavedAddress({required this.name, required this.ip, required this.port});

  Map<String, String> toJson() {
    return {
      'name': name,
      'ip': ip,
      'port': port,
    };
  }

  static SavedAddress fromJson(Map<String, String> json) {
    return SavedAddress(
      name: json['name']!,
      ip: json['ip']!,
      port: json['port']!,
    );
  }
}

class SavedAddressManager {
  Future<void> saveAddress(SavedAddress address) async {
    final prefs = await SharedPreferences.getInstance();
    final List<String> savedAddresses =
        prefs.getStringList('saved_addresses') ?? [];
    savedAddresses.add(address.toJson().toString());
    await prefs.setStringList('saved_addresses', savedAddresses);
  }

  Future<void> deleteAddress(SavedAddress address) async {
    final prefs = await SharedPreferences.getInstance();
    final List<String> savedAddresses =
        prefs.getStringList('saved_addresses') ?? [];
    savedAddresses
        .removeWhere((addr) => addr.contains('"name":"${address.name}"'));
    await prefs.setStringList('saved_addresses', savedAddresses);
  }

  Future<List<SavedAddress>> loadAddresses() async {
    final prefs = await SharedPreferences.getInstance();
    final List<String>? savedAddresses = prefs.getStringList('saved_addresses');

    if (savedAddresses != null) {
      return savedAddresses.map((addressString) {
        final addressJson = Map<String, String>.fromEntries(
          addressString
              .substring(1, addressString.length - 1)
              .split(', ')
              .map((e) {
            final pair = e.split(': ');
            return MapEntry(pair[0], pair[1]);
          }),
        );
        return SavedAddress.fromJson(addressJson);
      }).toList();
    } else {
      return [];
    }
  }
}
