import subprocess

# runAndExpect("rio asldfkhsldkfj", 3)
# output = runAndExpect("rio ps")

def run(cmd, status=0):
    # @TODO actually check status
    result = subprocess.check_output(cmd, shell=True)
    result = result.decode("utf-8").strip()
    
    return result