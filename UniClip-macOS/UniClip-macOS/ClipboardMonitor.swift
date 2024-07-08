import Foundation
import AppKit

class ClipboardMonitor: ObservableObject {
    @Published var isMonitoring = false
    private var timer: Timer?
    
    func toggleMonitoring() {
        if isMonitoring {
            stopMonitoring()
        } else {
            startMonitoring()
        }
    }
    
    private func startMonitoring() {
        isMonitoring = true
        timer = Timer.scheduledTimer(withTimeInterval: 1.0, repeats: true) { _ in
            self.checkClipboard()
        }
    }
    
    private func stopMonitoring() {
        isMonitoring = false
        timer?.invalidate()
        timer = nil
    }
    
    private func checkClipboard() {
        if let clipboardContent = NSPasteboard.general.string(forType: .string) {
            postClipboardData(data: clipboardContent)
        }
    }
    
    private func postClipboardData(data: String) {
        guard let url = URL(string: "http://127.0.0.1:3000/clipboard") else { return }
        
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        
        let payload: [String: String] = ["clipboardData": data]
        request.httpBody = try? JSONSerialization.data(withJSONObject: payload, options: [])
        
        let task = URLSession.shared.dataTask(with: request) { (data, response, error) in
            if let error = error {
                print("Error posting clipboard data: \(error)")
                return
            }
            
            guard let httpResponse = response as? HTTPURLResponse, (200...299).contains(httpResponse.statusCode) else {
                print("Server error")
                return
            }
            
            if let data = data, let responseString = String(data: data, encoding: .utf8) {
                print("Response: \(responseString)")
            }
        }
        task.resume()
    }
}
