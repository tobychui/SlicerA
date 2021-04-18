![](img/banner.png)

# SlicerA

A web based STL to Gcode slicer for ArozOS

## Installation

### Requirement

- Go 1.15 or above
- Debian Buster on ARM, x64 or Windows on x64 platforms
- [ArozOS](https://github.com/tobychui/arozos) v1.111 or above

### Build

1. Clone this repo into your ArozOS subservice directory (usually can be found under ~/arozos/subservice). 

   ```
   cd ~/arozos/subservice/
   git clone https://github.com/tobychui/SlicerA
   cd SlicerA
   ```

2. Build the SlicerA subservice using the build.sh bash script

   ```
   ./build.sh
   
   # Optional, depends on your permission settings
   sudo chmod 755 -R ./
   ```

3. Restart arozos  using systemctl 

   ```
   sudo systemctl restart arozos
   ```



## Usage

To use SlicerA, you can first upload some STL files to your ArozOS cloud desktop and follow the steps below

1. Load STL Model using the top right hand corner button or the "1. Load STL Model " button
2. Click "Slice to Gcode". Wait until it complete and check the finished gcode for any issues in slicing
3. Click "Save to File" if the gcode file looks good.



## Screenshots

![](img/1.png)



![](img/2.png)



![](img/3.png)



![](img/4.png)



And after export, you can see your gcode file in the location you selected.

![](img/6.png)



## License

Please see the LICENSE file



### Special Thanks

This project is powered by the amazing Golang written STL to Gcode slicer named [GoSlice](https://github.com/aligator/GoSlice)

